package websocket

import (
	"encoding/json"
	"log"
	"sync"
)

type Manager struct {
	mu      sync.RWMutex
	clients map[*Client]struct{}
	logger  *log.Logger
}

func NewManager(logger *log.Logger) *Manager {
	return &Manager{
		clients: make(map[*Client]struct{}),
		logger:  logger,
	}
}

func (m *Manager) Register(client *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clients[client] = struct{}{}
	m.logger.Printf("websocket client registered: %s", client.remoteAddr())
}

func (m *Manager) Unregister(client *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.clients[client]; !ok {
		return
	}
	delete(m.clients, client)
	close(client.send)
	m.logger.Printf("websocket client unregistered: %s", client.remoteAddr())
}

func (m *Manager) Broadcast(message []byte) {
	m.mu.RLock()
	clients := make([]*Client, 0, len(m.clients))
	for client := range m.clients {
		clients = append(clients, client)
	}
	clientCount := len(clients)
	m.mu.RUnlock()

	m.logger.Printf("broadcasting activity to %d websocket clients", clientCount)

	// Parse the message to identify the target tenant
	// If you later add tenant_id to the notification payload,
	// this filtering will work. Currently, it allows all messages through.
	var event struct {
		TenantID string `json:"tenant_id"`
	}
	if err := json.Unmarshal(message, &event); err != nil {
		// Message might not be valid JSON or missing tenant_id, we still broadcast
	}

	for _, client := range clients {
		// Filter: only broadcast if message is global or tenant matches
		if event.TenantID != "" && client.tenant != event.TenantID {
			continue
		}

		select {
		case client.send <- message:
		default:
			m.logger.Printf("websocket client send queue full, disconnecting: %s", client.remoteAddr())
			client.close()
			m.Unregister(client)
		}
	}
}

func (m *Manager) Shutdown() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for client := range m.clients {
		close(client.send)
		client.close()
		delete(m.clients, client)
	}
	m.logger.Println("websocket manager shutdown complete")
}
