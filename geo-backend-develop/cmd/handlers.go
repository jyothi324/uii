package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	hub *Hub

	upgrader = websocket.Upgrader{
		CheckOrigin:     websocketCheckOrigin,
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	mapboxSuggestLimiter  = newIPRateLimiter(30, time.Minute)
	mapboxRetrieveLimiter = newIPRateLimiter(15, time.Minute)
)

func (dp *DeliveryPerson) handleMessage(msg Message) {
	dp.mu.Lock()
	dp.LastSeen = time.Now()
	dp.mu.Unlock()

	switch msg.Type {
	case "ping":
		hub.sendToDelivery(dp, Message{Type: "pong", Timestamp: time.Now().Unix()})
	case "updateLocation":
		dp.handleLocationUpdate(msg)
	case "updateStatus":
		dp.handleStatusUpdate(msg)
	case "taskAction":
		dp.handleTaskAction(msg)
	case "updateConsent":
		dp.handleConsentUpdate(msg)
	}
}

func (dp *DeliveryPerson) handleLocationUpdate(msg Message) {
	dataMap, ok := msg.Data.(map[string]interface{})
	if !ok {
		return
	}
	coordsData, ok := dataMap["coords"].(map[string]interface{})
	if !ok {
		return
	}

	lat, _ := coordsData["latitude"].(float64)
	lng, _ := coordsData["longitude"].(float64)

	newCoords := &Coords{
		Latitude:  lat,
		Longitude: lng,
	}

	dp.mu.Lock()
	dp.CurrentLocation = newCoords
	assignedID := dp.AssignedTaskID
	allowTracking := dp.AllowAdminTracking
	dp.mu.Unlock()

	if allowTracking {
		hub.broadcastToAdmins(Message{
			Type: "agentLocation",
			Data: map[string]interface{}{
				"agentId":  dp.ID,
				"name":     dp.Name,
				"location": newCoords,
				"status":   dp.Status,
				"taskId":   dp.AssignedTaskID,
			},
			Timestamp: time.Now().Unix(),
		})
	}
	if assignedID != "" {
		hub.updateTaskLocation(assignedID, newCoords)
	}
}

func (dp *DeliveryPerson) handleStatusUpdate(msg Message) {
	dataMap, ok := msg.Data.(map[string]interface{})
	if !ok {
		return
	}
	status, _ := dataMap["status"].(string)
	dp.mu.Lock()
	dp.Status = status
	dp.mu.Unlock()
	hub.sendToDelivery(dp, Message{
		Type:      "statusUpdated",
		Data:      map[string]interface{}{"success": true, "status": status},
		Timestamp: time.Now().Unix(),
	})
}

func (dp *DeliveryPerson) handleTaskAction(msg Message) {
	dataMap, ok := msg.Data.(map[string]interface{})
	if !ok {
		return
	}
	actionRaw, ok := dataMap["action"]
	if !ok {
		return
	}
	action := strings.TrimSpace(fmt.Sprintf("%v", actionRaw))

	hub.mu.Lock()
	task, exists := hub.deliveryTasks[dp.AssignedTaskID]
	if !exists {
		hub.mu.Unlock()
		return
	}

	now := time.Now()
	var shouldAssignNext bool
	taskID := task.ID

	switch action {
	case "start_delivery":
		task.Status = "assigned"
	case "pickup":
		task.Status = "picked_up"
		task.PickedUpAt = &now
	case "start_transit":
		task.Status = "in_transit"
	case "deliver":
		task.Status = "delivered"
		task.DeliveredAt = &now
		task.RouteCoords = nil
		shouldAssignNext = true
		dp.mu.Lock()
		dp.AssignedTaskID = ""
		dp.Status = "available"
		dp.mu.Unlock()
	case "cancel":
		task.Status = "cancelled"
		task.RouteCoords = nil
		shouldAssignNext = true
		dp.mu.Lock()
		dp.AssignedTaskID = ""
		dp.Status = "available"
		dp.mu.Unlock()
	default:
		log.Printf("⚠️ Unknown action '%s' for task %s", action, taskID)
	}
	hub.mu.Unlock()

	hub.sendTrackingUpdate(taskID)

	cleanTask := *task
	cleanTask.RouteCoords = nil
	hub.sendToDelivery(dp, Message{
		Type: "taskActionSuccess",
		Data: map[string]interface{}{
			"success": true,
			"action":  action,
			"task":    cleanTask,
		},
		Timestamp: time.Now().Unix(),
	})

	if shouldAssignNext {
		hub.assignPendingTaskForAgent(dp)
	}
}

func (dp *DeliveryPerson) handleConsentUpdate(msg Message) {
	dataMap, ok := msg.Data.(map[string]interface{})
	if !ok {
		return
	}
	allow, _ := dataMap["allowTracking"].(bool)

	dp.mu.Lock()
	dp.AllowAdminTracking = allow
	dp.mu.Unlock()

	if !allow {
		hub.broadcastToAdmins(Message{
			Type:      "agentOffline",
			Data:      map[string]interface{}{"agentId": dp.ID},
			Timestamp: time.Now().Unix(),
		})
	}
}

func (dp *DeliveryPerson) readPump() {
	defer func() {
		if !hub.shuttingDown.Load() {
			hub.unregister <- dp
		}
		dp.Conn.Close()
	}()

	dp.Conn.SetReadLimit(2048)
	dp.Conn.SetReadDeadline(time.Now().Add(120 * time.Second))
	dp.Conn.SetPongHandler(func(string) error {
		dp.Conn.SetReadDeadline(time.Now().Add(120 * time.Second))
		dp.mu.Lock()
		dp.LastSeen = time.Now()
		dp.mu.Unlock()
		return nil
	})

	for {
		var msg Message
		err := dp.Conn.ReadJSON(&msg)
		if err != nil {
			break
		}
		dp.handleMessage(msg)
	}
}

func (dp *DeliveryPerson) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		dp.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-dp.Send:
			dp.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				dp.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := dp.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			dp.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := dp.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (tc *TrackingClient) readPump() {
	defer func() {
		if !hub.shuttingDown.Load() {
			hub.unregister <- tc
		}
		tc.Conn.Close()
	}()

	tc.Conn.SetReadLimit(512)
	tc.Conn.SetReadDeadline(time.Now().Add(120 * time.Second))
	tc.Conn.SetPongHandler(func(string) error {
		tc.Conn.SetReadDeadline(time.Now().Add(120 * time.Second))
		tc.mu.Lock()
		tc.LastSeen = time.Now()
		tc.mu.Unlock()
		return nil
	})

	for {
		var msg Message
		if err := tc.Conn.ReadJSON(&msg); err != nil {
			break
		}
		tc.mu.Lock()
		tc.LastSeen = time.Now()
		tc.mu.Unlock()

		if msg.Type == "ping" {
			data, _ := json.Marshal(Message{Type: "pong", Timestamp: time.Now().Unix()})
			tc.Send <- data
		} else if msg.Type == "requestUpdate" {
			hub.sendTrackingUpdate(tc.TaskID)
		}
	}
}

func (tc *TrackingClient) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		tc.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-tc.Send:
			tc.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				tc.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := tc.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			tc.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := tc.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (ac *AdminClient) readPump() {
	defer func() {
		ac.Conn.Close()
	}()

	ac.Conn.SetReadLimit(1024)
	ac.Conn.SetReadDeadline(time.Now().Add(120 * time.Second))
	ac.Conn.SetPongHandler(func(string) error {
		ac.Conn.SetReadDeadline(time.Now().Add(120 * time.Second))
		return nil
	})

	for {
		_, data, err := ac.Conn.ReadMessage()
		if err != nil {
			break
		}
		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}
		if msg.Type == "requestSnapshot" {
			hub.mu.RLock()
			for _, dp := range hub.deliveryPersons {
				dp.mu.RLock()
				allow := dp.AllowAdminTracking
				snap := Message{Type: "agentLocation",
					Data: map[string]interface{}{
						"agentId":  dp.ID,
						"name":     dp.Name,
						"location": dp.CurrentLocation,
						"status":   dp.Status,
						"taskId":   dp.AssignedTaskID,
					},
					Timestamp: time.Now().Unix()}
				dp.mu.RUnlock()

				if !allow {
					continue
				}

				snapData, _ := json.Marshal(snap)
				select {
				case ac.Send <- snapData:
				default:
				}
			}
			hub.mu.RUnlock()
		}
	}
}

func (ac *AdminClient) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		ac.Conn.Close()
		hub.mu.Lock()
		delete(hub.adminClients, ac.ID)
		hub.mu.Unlock()
	}()

	for {
		select {
		case message, ok := <-ac.Send:
			ac.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				ac.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := ac.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			ac.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := ac.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func rejectIfShuttingDown(c *gin.Context) bool {
	if hub.shuttingDown.Load() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "server is shutting down"})
		return true
	}
	return false
}

func handleDeliveryWS(c *gin.Context) {
	if rejectIfShuttingDown(c) {
		return
	}
	agentID := c.Query("agentId")
	if agentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agentId is required"})
		return
	}

	if !authDisabled() {
		authenticatedID := getRequestUserID(c)
		if authenticatedID == "" || authenticatedID != agentID {
			c.JSON(http.StatusForbidden, gin.H{"error": "agentId does not match authenticated user"})
			return
		}
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	name := c.Query("name")
	vehicleType := c.Query("vehicleType")
	vehicleNumber := c.Query("vehicleNumber")

	if name == "" {
		name = "Delivery Person"
	}
	if vehicleType == "" {
		vehicleType = "bike"
	}

	dp := &DeliveryPerson{
		ID:            agentID,
		Name:          name,
		VehicleType:   vehicleType,
		VehicleNumber: vehicleNumber,
		Status:        "available",
		Conn:          conn,
		Send:          make(chan []byte, 256),
		LastSeen:      time.Now(),
	}

	hub.register <- dp
	go dp.writePump()
	go dp.readPump()
}

func handleTrackingWS(c *gin.Context) {
	if rejectIfShuttingDown(c) {
		return
	}
	taskID := c.Param("taskId")
	hub.mu.RLock()
	_, exists := hub.deliveryTasks[taskID]
	hub.mu.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	tc := &TrackingClient{
		ID:       generateID("TRACK"),
		TaskID:   taskID,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		LastSeen: time.Now(),
	}

	hub.register <- tc
	go tc.writePump()
	go tc.readPump()
}

func handleAdminWS(c *gin.Context) {
	if rejectIfShuttingDown(c) {
		return
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	ac := &AdminClient{
		ID:   generateID("ADMIN"),
		Conn: conn,
		Send: make(chan []byte, 256),
	}

	hub.mu.Lock()
	hub.adminClients[ac.ID] = ac
	hub.mu.Unlock()

	hub.mu.RLock()
	for _, dp := range hub.deliveryPersons {
		dp.mu.RLock()
		loc := dp.CurrentLocation
		allow := dp.AllowAdminTracking
		dp.mu.RUnlock()

		if loc == nil || !allow {
			continue
		}

		msg := Message{
			Type: "agentLocation",
			Data: map[string]interface{}{
				"agentId":  dp.ID,
				"name":     dp.Name,
				"location": loc,
				"status":   dp.Status,
				"taskId":   dp.AssignedTaskID,
			},
			Timestamp: time.Now().Unix(),
		}
		data, _ := json.Marshal(msg)
		select {
		case ac.Send <- data:
		default:
		}
	}
	hub.mu.RUnlock()

	go ac.writePump()
	go ac.readPump()
}

func handleCreateTask(c *gin.Context) {
	taskID := c.Param("id")
	var req Address
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format. Expected {agentId, address}"})
		return
	}

	if strings.TrimSpace(req.AgentID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agentId is required"})
		return
	}

	agentID := strings.TrimSpace(req.AgentID)
	hub.mu.RLock()
	dp, agentOnline := hub.deliveryPersons[agentID]
	hub.mu.RUnlock()

	if agentOnline {
		dp.mu.RLock()
		dpStatus := dp.Status
		dpAssigned := dp.AssignedTaskID
		dp.mu.RUnlock()
		if dpStatus == "busy" {
			c.JSON(http.StatusConflict, gin.H{
				"error":         "Agent is currently busy with another delivery.",
				"agentId":       agentID,
				"currentTaskId": dpAssigned,
				"status":        "rejected",
			})
			return
		}
	}

	coords, err := geocodeAddress(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to geocode address"})
		return
	}

	baseURL := "https://app.tulipcrm.com"
	trackingLink := fmt.Sprintf("%s/track?id=%s", baseURL, taskID)

	task := &DeliveryTask{
		ID:                taskID,
		AgentID:           agentID,
		DestinationCoords: coords,
		DestinationAddr:   &req,
		Status:            "pending",
		CreatedAt:         time.Now(),
		TrackingLink:      trackingLink,
		RouteCoords:       []Coords{},
	}

	hub.mu.Lock()
	hub.deliveryTasks[taskID] = task
	hub.mu.Unlock()

	if agentOnline && dp != nil {
		dp.mu.RLock()
		isAvailable := dp.Status == "available"
		dp.mu.RUnlock()
		if isAvailable {
			hub.assignPendingTaskForAgent(dp)
		} else {
			hub.sendToDelivery(dp, Message{
				Type: "taskPendingWait",
				Data: map[string]interface{}{
					"message": "Please complete your current delivery before starting a new one.",
					"task":    task,
				},
				Timestamp: time.Now().Unix(),
			})
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":      true,
		"task":         task,
		"trackingLink": trackingLink,
	})
}

func handleGetTask(c *gin.Context) {
	taskID := c.Param("taskId")
	hub.mu.RLock()
	task, exists := hub.deliveryTasks[taskID]
	hub.mu.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "task": task})
}

func handleGetAllTasks(c *gin.Context) {
	hub.mu.RLock()
	tasks := make([]*DeliveryTask, 0, len(hub.deliveryTasks))
	for _, task := range hub.deliveryTasks {
		tasks = append(tasks, task)
	}
	hub.mu.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"count":   len(tasks),
		"tasks":   tasks,
	})
}

func handleGetStats(c *gin.Context) {
	hub.mu.RLock()
	deliveryList := make([]map[string]interface{}, 0)

	for _, dp := range hub.deliveryPersons {
		dp.mu.RLock()
		deliveryList = append(deliveryList, map[string]interface{}{
			"id":             dp.ID,
			"name":           dp.Name,
			"status":         dp.Status,
			"location":       dp.CurrentLocation,
			"assignedTaskId": dp.AssignedTaskID,
		})
		dp.mu.RUnlock()
	}

	totalTasks := len(hub.deliveryTasks)
	totalTrackers := len(hub.trackingClients)
	hub.mu.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"totalDeliveries": len(deliveryList),
		"totalTasks":      totalTasks,
		"totalTrackers":   totalTrackers,
		"deliveryPersons": deliveryList,
		"timestamp":       time.Now().Unix(),
	})
}

func handleNotifyCustomer(c *gin.Context) {
	var req NotifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}

	if strings.TrimSpace(req.Phone) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "phone is required"})
		return
	}

	trackingURL := strings.TrimSpace(req.TrackingURL)
	if trackingURL == "" {
		orderID := strings.TrimSpace(req.OrderID)
		if orderID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "orderId required when trackingUrl is not provided"})
			return
		}
		trackingURL = fmt.Sprintf("%s/r/%s", getPublicBaseURL(c), orderID)
	} else {
		if !strings.HasPrefix(trackingURL, "http://") && !strings.HasPrefix(trackingURL, "https://") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "trackingUrl must start with http:// or https://"})
			return
		}
	}

	body := strings.TrimSpace(req.Message)
	if body == "" {
		body = trackingSMSBody(trackingURL, req.OrderID)
	}

	if strings.TrimSpace(req.OrderID) != "" {
		orderID := strings.TrimSpace(req.OrderID)
		go func(taskID string) {
			hub.mu.RLock()
			task, exists := hub.deliveryTasks[taskID]
			assignedDPID := ""
			agentID := ""
			if exists {
				assignedDPID = task.AssignedToID
				agentID = task.AgentID
			}
			hub.mu.RUnlock()
			if !exists {
				return
			}

			var dp *DeliveryPerson
			if assignedDPID != "" {
				hub.mu.RLock()
				dp, _ = hub.deliveryPersons[assignedDPID]
				hub.mu.RUnlock()
			}
			if dp == nil && agentID != "" {
				hub.mu.RLock()
				dp, _ = hub.deliveryPersons[agentID]
				hub.mu.RUnlock()
			}
			if dp == nil {
				return
			}

			dp.mu.RLock()
			curLoc := dp.CurrentLocation
			dp.mu.RUnlock()
			if curLoc == nil {
				return
			}

			hub.mu.RLock()
			destLoc := task.DestinationCoords
			hub.mu.RUnlock()
			if destLoc == nil {
				return
			}

			profile := defaultProfile
			dp.mu.RLock()
			if dp.VehicleType != "" {
				vt := strings.ToLower(dp.VehicleType)
				if p, ok := vehicleProfileMap[vt]; ok {
					profile = p
				}
			}
			dp.mu.RUnlock()

			route, distanceMiles, durationHours, err := fetchRoute(
				curLoc.Latitude, curLoc.Longitude,
				destLoc.Latitude, destLoc.Longitude,
				profile,
			)

			if err != nil {
				route = []Coords{*curLoc, *destLoc}
				distanceMiles = calculateDistance(curLoc.Latitude, curLoc.Longitude, destLoc.Latitude, destLoc.Longitude)
				durationHours = distanceMiles / 37.28
			}

			eta := formatHours(durationHours)

			hub.mu.Lock()
			if t, ok := hub.deliveryTasks[taskID]; ok {
				t.CurrentLocation = curLoc
				t.RouteCoords = route
				t.Distance = distanceMiles
				t.ETA = eta
			}
			hub.mu.Unlock()

			hub.sendTrackingUpdate(taskID)

			hub.mu.RLock()
			updatedTask := hub.deliveryTasks[taskID]
			hub.mu.RUnlock()

			hub.sendToDelivery(dp, Message{Type: "routeUpdated", Data: updatedTask, Timestamp: time.Now().Unix()})
			hub.sendToDelivery(dp, Message{Type: "taskAssigned", Data: updatedTask, Timestamp: time.Now().Unix()})
		}(orderID)
	}

	tc := loadTwilioConfig()
	if !tc.IsConfigured() {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Twilio not configured on server"})
		return
	}
	if err := tc.SendSMS(req.Phone, body); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to send SMS", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      "sent",
		"sentTo":      safePhone(req.Phone),
		"trackingUrl": trackingURL,
		"orderId":     req.OrderID,
	})
}

func handleMessageNotification(c *gin.Context) {
	var req MsgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
		return
	}
	tc := loadTwilioConfig()
	if err := tc.SendSMS(req.Phone, req.Body); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to send SMS", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "sent"})
}

func handleTrackingPage(c *gin.Context) {
	taskID := c.Param("taskId")
	hub.mu.RLock()
	task, exists := hub.deliveryTasks[taskID]
	hub.mu.RUnlock()

	if !exists {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"message": "Task not found"})
		return
	}

	c.HTML(http.StatusOK, "tracking.html", gin.H{
		"taskId": taskID,
		"task":   task,
		"wsURL":  fmt.Sprintf("ws://%s/ws/track/%s", c.Request.Host, taskID),
	})
}

func handleTrackingRedirect(c *gin.Context) {
	taskID := c.Param("taskId")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "taskId required"})
		return
	}
	if strings.TrimSpace(trackingBase) == "" {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "TRACKING_BASE_URL not configured on backend"})
		return
	}
	c.Redirect(http.StatusTemporaryRedirect, buildTrackingLink(c, taskID))
}

type mockLocationAddress struct {
	ID        int     `json:"id"`
	State     string  `json:"state"`
	County    string  `json:"county"`
	Street    string  `json:"street"`
	UnitSuite string  `json:"unit_suite"`
	City      string  `json:"city"`
	Zip       string  `json:"zip"`
	Formatted string  `json:"formatted"`
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
}

func isLocationMockEnabled(c *gin.Context) bool {
	headerValue := strings.TrimSpace(c.GetHeader("X-Location-Mock"))
	if headerValue == "" {
		return false
	}
	return strings.EqualFold(headerValue, "true") || headerValue == "1"
}

func loadMockLocationAddresses() ([]mockLocationAddress, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	paths := []string{
		filepath.Join(cwd, "mock", "addresses.json"),
		filepath.Join(cwd, "mock", "mock_addresses_db_schema.json"),
		filepath.Join(cwd, "..", "mock", "addresses.json"),
		filepath.Join(cwd, "..", "mock", "mock_addresses_db_schema.json"),
	}

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var addresses []mockLocationAddress
		if err := json.Unmarshal(data, &addresses); err != nil {
			return nil, err
		}
		log.Printf("🧭 [MOCK LOCATION] loaded %d addresses from %s", len(addresses), path)
		return addresses, nil
	}

	return nil, fmt.Errorf("mock location addresses file not found")
}

func findMockLocationMatches(query string) ([]mockLocationAddress, error) {
	addresses, err := loadMockLocationAddresses()
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(query) == "" {
		return nil, nil
	}

	searchTerm := strings.ToLower(strings.TrimSpace(query))
	matches := make([]mockLocationAddress, 0)
	for _, address := range addresses {
		candidate := strings.ToLower(strings.Join([]string{
			address.Street,
			address.UnitSuite,
			address.City,
			address.State,
			address.County,
			address.Zip,
			address.Formatted,
		}, " "))
		if strings.Contains(candidate, searchTerm) {
			matches = append(matches, address)
		}
	}
	return matches, nil
}

func buildMockLocationSuggestResponse(query string) (map[string]interface{}, error) {
	matches, err := findMockLocationMatches(query)
	if err != nil {
		return nil, err
	}

	suggestions := make([]map[string]interface{}, 0, len(matches))
	for _, address := range matches {
		suggestions = append(suggestions, map[string]interface{}{
			"mapbox_id":       strconv.Itoa(address.ID),
			"name":            address.Street,
			"place_formatted": address.Formatted,
			"address": map[string]interface{}{
				"street":   address.Street,
				"city":     address.City,
				"state":    address.State,
				"postcode": address.Zip,
				"county":   address.County,
			},
		})
	}

	return map[string]interface{}{"suggestions": suggestions}, nil
}

func buildMockLocationRetrieveResponse(mapboxID string) (map[string]interface{}, error) {
	addresses, err := loadMockLocationAddresses()
	if err != nil {
		return nil, err
	}

	id, err := strconv.Atoi(strings.TrimSpace(mapboxID))
	if err != nil {
		return map[string]interface{}{"features": []interface{}{}}, nil
	}

	for _, address := range addresses {
		if address.ID != id {
			continue
		}
		feature := map[string]interface{}{
			"type": "Feature",
			"properties": map[string]interface{}{
				"mapbox_id":       strconv.Itoa(address.ID),
				"name":            address.Street,
				"place_formatted": address.Formatted,
				"context": map[string]interface{}{
					"address":  map[string]interface{}{"name": address.Street},
					"place":    map[string]interface{}{"name": address.City},
					"locality": map[string]interface{}{"name": address.City},
					"region":   map[string]interface{}{"name": address.State},
					"postcode": map[string]interface{}{"name": address.Zip},
					"country":  map[string]interface{}{"name": "United States"},
				},
			},
			"geometry": map[string]interface{}{
				"type":        "Point",
				"coordinates": []float64{address.Lng, address.Lat},
			},
		}
		return map[string]interface{}{"features": []interface{}{feature}}, nil
	}

	return map[string]interface{}{"features": []interface{}{}}, nil
}

func handleMapboxSuggest(c *gin.Context) {
	clientIP := c.ClientIP()
	if !mapboxSuggestLimiter.allow(clientIP) {
		c.JSON(429, gin.H{"error": "rate limit exceeded"})
		return
	}

	query := strings.TrimSpace(c.Query("q"))
	sessionToken := c.Query("session_token")

	if query == "" {
		c.JSON(400, gin.H{"error": "query is required"})
		return
	}
	if len(query) < 5 {
		c.JSON(400, gin.H{"error": "query must be at least 5 characters"})
		return
	}
	if sessionToken == "" {
		c.JSON(400, gin.H{"error": "session_token is required"})
		return
	}
	if isLocationMockEnabled(c) {
		payload, err := buildMockLocationSuggestResponse(query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "mock location search failed"})
			return
		}
		c.JSON(http.StatusOK, payload)
		return
	}

	mapboxToken := getMapboxToken()
	if mapboxToken == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Mapbox is not configured"})
		return
	}

	suggestURL := fmt.Sprintf(
		"https://api.mapbox.com/search/searchbox/v1/suggest?q=%s&types=address&country=US&session_token=%s&access_token=%s",
		url.QueryEscape(query),
		url.QueryEscape(sessionToken),
		mapboxToken,
	)

	resp, err := http.Get(suggestURL)
	if err != nil {
		c.JSON(500, gin.H{"error": "suggest failed"})
		return
	}
	defer resp.Body.Close()

	var data interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		c.JSON(500, gin.H{"error": "invalid suggest response"})
		return
	}
	c.JSON(200, data)
}

func handleMapboxRetrieve(c *gin.Context) {
	clientIP := c.ClientIP()
	if !mapboxRetrieveLimiter.allow(clientIP) {
		c.JSON(429, gin.H{"error": "rate limit exceeded"})
		return
	}

	mapboxID := c.Query("mapbox_id")
	sessionToken := c.Query("session_token")
	if mapboxID == "" {
		c.JSON(400, gin.H{"error": "mapbox_id is required"})
		return
	}
	if sessionToken == "" {
		c.JSON(400, gin.H{"error": "session_token is required"})
		return
	}
	if isLocationMockEnabled(c) {
		payload, err := buildMockLocationRetrieveResponse(mapboxID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "mock location retrieve failed"})
			return
		}
		c.JSON(http.StatusOK, payload)
		return
	}

	mapboxToken := getMapboxToken()
	if mapboxToken == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Mapbox is not configured"})
		return
	}

	retrieveURL := fmt.Sprintf(
		"https://api.mapbox.com/search/searchbox/v1/retrieve/%s?session_token=%s&access_token=%s",
		url.QueryEscape(mapboxID),
		url.QueryEscape(sessionToken),
		mapboxToken,
	)

	resp, err := http.Get(retrieveURL)
	if err != nil {
		c.JSON(500, gin.H{"error": "retrieve request failed"})
		return
	}
	defer resp.Body.Close()

	var data interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		c.JSON(500, gin.H{"error": "invalid response"})
		return
	}
	c.JSON(200, data)
}
