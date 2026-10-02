package main

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

type Address struct {
	ID        uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	State     string `json:"state" gorm:"type:text;not null"`
	County    string `json:"county" gorm:"type:text;not null"`
	Street    string `json:"street" gorm:"type:varchar(255);not null"`
	UnitSuite string `json:"unitSuite" gorm:"type:varchar(255);not null"`
	City      string `json:"city" gorm:"type:varchar(255);not null"`
	Zip       string `json:"zip" gorm:"default:false"`
	AgentID   string `json:"agentId" gorm:"type:varchar(255);not null"`
}

type AdminClient struct {
	ID   string
	Conn *websocket.Conn
	Send chan []byte
	mu   sync.RWMutex
}

type Coords struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"address,omitempty"`
}

type NominatimResponse struct {
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
	DisplayName string `json:"display_name"`
}

type DeliveryTask struct {
	ID                    string     `json:"id"`
	DestinationCoords     *Coords    `json:"destinationCoords"`
	DestinationAddr       *Address   `json:"destinationAddress,omitempty"`
	Status                string     `json:"status"`
	AssignedToID          string     `json:"assignedToId,omitempty"`
	AgentID               string     `json:"agentId,omitempty"`
	CreatedAt             time.Time  `json:"createdAt"`
	AssignedAt            *time.Time `json:"assignedAt,omitempty"`
	PickedUpAt            *time.Time `json:"pickedUpAt,omitempty"`
	DeliveredAt           *time.Time `json:"deliveredAt,omitempty"`
	TrackingLink          string     `json:"trackingLink"`
	CurrentLocation       *Coords    `json:"currentLocation,omitempty"`
	RouteCoords           []Coords   `json:"routeCoords,omitempty"`
	Distance              float64    `json:"distance,omitempty"`
	ETA                   string     `json:"eta,omitempty"`
	LastRouteCalcAt       *time.Time `json:"-"`
	LastRouteCalcLocation *Coords    `json:"-"`
}

type DeliveryPerson struct {
	ID                 string          `json:"id"`
	Name               string          `json:"name"`
	VehicleType        string          `json:"vehicleType"`
	VehicleNumber      string          `json:"vehicleNumber"`
	CurrentLocation    *Coords         `json:"currentLocation"`
	Status             string          `json:"status"`
	AssignedTaskID     string          `json:"assignedTaskId,omitempty"`
	Conn               *websocket.Conn `json:"-"`
	Send               chan []byte     `json:"-"`
	LastSeen           time.Time       `json:"-"`
	mu                 sync.RWMutex    `json:"-"`
	AllowAdminTracking bool            `json:"-"`
}

type TrackingClient struct {
	ID       string
	TaskID   string
	Conn     *websocket.Conn
	Send     chan []byte
	LastSeen time.Time
	mu       sync.RWMutex
}

type Hub struct {
	deliveryPersons map[string]*DeliveryPerson
	deliveryTasks   map[string]*DeliveryTask
	trackingClients map[string]*TrackingClient
	register        chan interface{}
	unregister      chan interface{}
	stop            chan struct{}
	mu              sync.RWMutex
	adminClients    map[string]*AdminClient
	shuttingDown    atomic.Bool
}

type Message struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

type TwilioConfig struct {
	AccountSID string
	AuthToken  string
	From       string
}

type NotifyRequest struct {
	Phone       string `json:"phone"`
	TrackingURL string `json:"trackingUrl,omitempty"`
	OrderID     string `json:"orderId,omitempty"`
	Message     string `json:"message,omitempty"`
	Resend      bool   `json:"resend,omitempty"`
}

type MsgRequest struct {
	Body  string `json:"body"`
	Phone string `json:"phone"`
}

type CreateTaskRequest struct {
	Address Address `json:"address"`
	AgentID string  `json:"agentId"`
}

type ipRateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

func (dp *DeliveryPerson) Lock()    { dp.mu.Lock() }
func (dp *DeliveryPerson) Unlock()  { dp.mu.Unlock() }
func (dp *DeliveryPerson) RLock()   { dp.mu.RLock() }
func (dp *DeliveryPerson) RUnlock() { dp.mu.RUnlock() }

func (tc *TrackingClient) Lock()    { tc.mu.Lock() }
func (tc *TrackingClient) Unlock()  { tc.mu.Unlock() }
func (tc *TrackingClient) RLock()   { tc.mu.RLock() }
func (tc *TrackingClient) RUnlock() { tc.mu.RUnlock() }

func (ac *AdminClient) Lock()    { ac.mu.Lock() }
func (ac *AdminClient) Unlock()  { ac.mu.Unlock() }
func (ac *AdminClient) RLock()   { ac.mu.RLock() }
func (ac *AdminClient) RUnlock() { ac.mu.RUnlock() }
