package main

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

func (h *Hub) run() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-h.stop:
			return
		case item := <-h.register:
			if h.shuttingDown.Load() {
				h.rejectRegister(item)
				continue
			}
			h.handleRegister(item)
		case item := <-h.unregister:
			h.handleUnregister(item)
		case <-ticker.C:
			if h.shuttingDown.Load() {
				continue
			}
			h.cleanupInactive()
		}
	}
}

func (h *Hub) rejectRegister(item interface{}) {
	switch v := item.(type) {
	case *DeliveryPerson:
		closeWebSocket(v.Conn, v.Send)
	case *TrackingClient:
		closeWebSocket(v.Conn, v.Send)
	}
}

func (h *Hub) Shutdown() {
	if h.shuttingDown.Swap(true) {
		return
	}

	h.mu.Lock()
	for _, dp := range h.deliveryPersons {
		closeWebSocket(dp.Conn, dp.Send)
	}
	for _, tc := range h.trackingClients {
		closeWebSocket(tc.Conn, tc.Send)
	}
	for _, ac := range h.adminClients {
		closeWebSocket(ac.Conn, ac.Send)
	}
	h.deliveryPersons = make(map[string]*DeliveryPerson)
	h.trackingClients = make(map[string]*TrackingClient)
	h.adminClients = make(map[string]*AdminClient)
	h.mu.Unlock()

	close(h.stop)
}

func closeWebSocket(conn *websocket.Conn, send chan []byte) {
	if conn != nil {
		closeMsg := websocket.FormatCloseMessage(websocket.CloseGoingAway, "server shutting down")
		_ = conn.WriteMessage(websocket.CloseMessage, closeMsg)
		_ = conn.Close()
	}
	safeCloseSend(send)
}

func safeCloseSend(send chan []byte) {
	if send == nil {
		return
	}
	defer func() { recover() }()
	close(send)
}

func (h *Hub) handleRegister(item interface{}) {
	switch v := item.(type) {
	case *DeliveryPerson:
		h.mu.Lock()
		h.deliveryPersons[v.ID] = v
		h.mu.Unlock()
		h.sendToDelivery(v, Message{
			Type: "connected",
			Data: map[string]interface{}{
				"success":    true,
				"deliveryId": v.ID,
				"name":       v.Name,
			},
			Timestamp: time.Now().Unix(),
		})
		h.assignPendingTaskForAgent(v)

	case *TrackingClient:
		h.mu.Lock()
		h.trackingClients[v.ID] = v
		h.mu.Unlock()
		h.sendTrackingUpdate(v.TaskID)
	}
}

func (h *Hub) handleUnregister(item interface{}) {
	switch v := item.(type) {
	case *DeliveryPerson:
		h.mu.Lock()
		delete(h.deliveryPersons, v.ID)
		h.mu.Unlock()
		safeCloseSend(v.Send)
		h.broadcastToAdmins(Message{
			Type:      "agentOffline",
			Data:      map[string]interface{}{"agentId": v.ID},
			Timestamp: time.Now().Unix(),
		})

	case *TrackingClient:
		h.mu.Lock()
		delete(h.trackingClients, v.ID)
		h.mu.Unlock()
		safeCloseSend(v.Send)
	}
}

func (h *Hub) cleanupInactive() {
	h.mu.Lock()
	defer h.mu.Unlock()

	timeout := time.Now().Add(-5 * time.Minute)

	for id, dp := range h.deliveryPersons {
		dp.mu.RLock()
		lastSeen := dp.LastSeen
		dp.mu.RUnlock()
		if lastSeen.Before(timeout) {
			delete(h.deliveryPersons, id)
			safeCloseSend(dp.Send)
		}
	}

	for id, tc := range h.trackingClients {
		tc.mu.RLock()
		lastSeen := tc.LastSeen
		tc.mu.RUnlock()
		if lastSeen.Before(timeout) {
			delete(h.trackingClients, id)
			safeCloseSend(tc.Send)
		}
	}
}

func (h *Hub) sendToDelivery(dp *DeliveryPerson, msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	select {
	case dp.Send <- data:
	case <-time.After(100 * time.Millisecond):
	}
}

func (h *Hub) sendTrackingUpdate(taskID string) {
	h.mu.RLock()
	task, exists := h.deliveryTasks[taskID]
	if !exists {
		h.mu.RUnlock()
		return
	}

	msg := Message{
		Type:      "trackingUpdate",
		Data:      task,
		Timestamp: time.Now().Unix(),
	}
	data, _ := json.Marshal(msg)

	for _, tc := range h.trackingClients {
		tc.mu.RLock()
		tcTaskID := tc.TaskID
		tc.mu.RUnlock()
		if tcTaskID == taskID {
			select {
			case tc.Send <- data:
			case <-time.After(50 * time.Millisecond):
			}
		}
	}
	h.mu.RUnlock()
}

func (h *Hub) broadcastToAdmins(msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, ac := range h.adminClients {
		select {
		case ac.Send <- data:
		case <-time.After(50 * time.Millisecond):
		}
	}
}

func (h *Hub) assignPendingTaskForAgent(dp *DeliveryPerson) {
	h.mu.Lock()

	dp.mu.RLock()
	dpStatus := dp.Status
	dpID := dp.ID
	dp.mu.RUnlock()

	if dpStatus != "available" {
		var pendingTask *DeliveryTask
		for _, t := range h.deliveryTasks {
			if t.AgentID == dpID && t.Status == "pending" {
				pendingTask = t
				break
			}
		}
		h.mu.Unlock()
		if pendingTask != nil {
			h.sendToDelivery(dp, Message{
				Type: "taskPendingWait",
				Data: map[string]interface{}{
					"message": "Please complete your current delivery before starting a new one.",
					"task":    pendingTask,
				},
				Timestamp: time.Now().Unix(),
			})
		}
		return
	}

	var targetTask *DeliveryTask
	for _, t := range h.deliveryTasks {
		if t.AgentID == dpID && t.Status == "pending" {
			if targetTask == nil || t.CreatedAt.Before(targetTask.CreatedAt) {
				targetTask = t
			}
		}
	}

	if targetTask == nil {
		h.mu.Unlock()
		return
	}

	now := time.Now()
	targetTask.Status = "assigned"
	targetTask.AssignedToID = dpID
	targetTask.AssignedAt = &now
	targetTask.Distance = 0
	targetTask.ETA = "Calculating..."
	targetTask.RouteCoords = nil

	dp.mu.Lock()
	dp.AssignedTaskID = targetTask.ID
	dp.Status = "busy"
	dp.mu.Unlock()

	destCopy := *targetTask.DestinationCoords
	var startCopy Coords
	dp.mu.RLock()
	if dp.CurrentLocation != nil {
		startCopy = *dp.CurrentLocation
	}
	dp.mu.RUnlock()

	taskID := targetTask.ID
	h.mu.Unlock()

	h.sendToDelivery(dp, Message{
		Type:      "taskAssigned",
		Data:      targetTask,
		Timestamp: time.Now().Unix(),
	})
	h.sendTrackingUpdate(taskID)

	go func(tID string, start, dest Coords) {
		profile := defaultProfile
		dp.mu.RLock()
		if dp.VehicleType != "" {
			vt := strings.ToLower(dp.VehicleType)
			if p, ok := vehicleProfileMap[vt]; ok {
				profile = p
			}
		}
		dp.mu.RUnlock()
		hub.applyRouteToTask(tID, start, dest, profile)
	}(taskID, startCopy, destCopy)
}

func (h *Hub) applyRouteToTask(taskID string, start, dest Coords, profile string) {
	route, distanceMiles, durationHours, err := fetchRoute(
		start.Latitude, start.Longitude,
		dest.Latitude, dest.Longitude,
		profile,
	)
	if err != nil {
		distanceMiles = calculateDistance(start.Latitude, start.Longitude, dest.Latitude, dest.Longitude)
		durationHours = distanceMiles / 37.28
		route = []Coords{start, dest}
	}

	eta := formatHours(durationHours)
	now := time.Now()
	startCopy := start

	h.mu.Lock()
	if t, ok := h.deliveryTasks[taskID]; ok {
		t.RouteCoords = route
		t.Distance = distanceMiles
		t.ETA = eta
		t.LastRouteCalcAt = &now
		t.LastRouteCalcLocation = &startCopy
	}
	h.mu.Unlock()

	h.sendTrackingUpdate(taskID)
}

func (h *Hub) updateTaskLocation(taskID string, location *Coords) {
	h.mu.Lock()
	task, exists := h.deliveryTasks[taskID]
	if !exists {
		h.mu.Unlock()
		return
	}

	task.CurrentLocation = location

	var shouldRecalc bool
	var startCopy, destCopy Coords
	profile := defaultProfile

	if task.DestinationCoords != nil {
		dest := task.DestinationCoords
		task.CurrentLocation = location
		if throttledDirectionsEnabled() && shouldRecalculateRoute(task, *location) {
			shouldRecalc = true
			startCopy = *location
			destCopy = *dest
			if task.AssignedToID != "" {
				if dp, ok := h.deliveryPersons[task.AssignedToID]; ok {
					dp.mu.RLock()
					if dp.VehicleType != "" {
						vt := strings.ToLower(dp.VehicleType)
						if p, ok := vehicleProfileMap[vt]; ok {
							profile = p
						}
					}
					dp.mu.RUnlock()
				}
			}
		}
	}

	h.mu.Unlock()
	h.sendTrackingUpdate(taskID)

	if task.AssignedToID != "" {
		h.mu.RLock()
		dp, exists := h.deliveryPersons[task.AssignedToID]
		h.mu.RUnlock()
		if exists {
			h.sendToDelivery(dp, Message{
				Type:      "trackingUpdate",
				Data:      task,
				Timestamp: time.Now().Unix(),
			})
		}
	}

	if shouldRecalc {
		go h.applyRouteToTask(taskID, startCopy, destCopy, profile)
	}
}

func shouldRecalculateRoute(task *DeliveryTask, location Coords) bool {
	if task.LastRouteCalcAt == nil || task.LastRouteCalcLocation == nil {
		return true
	}
	if time.Since(*task.LastRouteCalcAt) >= routeRecalcMinInterval() {
		return true
	}
	prev := task.LastRouteCalcLocation
	return calculateDistanceMeters(
		prev.Latitude, prev.Longitude,
		location.Latitude, location.Longitude,
	) >= routeRecalcMinMeters()
}
