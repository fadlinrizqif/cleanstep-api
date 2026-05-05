package ws

import (
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type PaymentEvent struct {
	OrderId uuid.UUID
	UserId  uuid.UUID
	Status  string
}

type Hub struct {
	Clients map[uuid.UUID]*websocket.Conn
	EventCh chan PaymentEvent
	Mu      *sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		Clients: make(map[uuid.UUID]*websocket.Conn),
		EventCh: make(chan PaymentEvent),
	}
}

func (h *Hub) Run() {
	for event := range h.EventCh {
		conn, exist := h.Clients[event.UserId]
		if !exist {
			continue
		}

		conn.WriteJSON(map[string]string{
			"event":    "payment_success",
			"order_id": event.OrderId.String(),
			"status":   event.Status,
		})
	}
}

func (h *Hub) Register(userID uuid.UUID, conn *websocket.Conn) {
	h.Mu.Lock()
	defer h.Mu.Unlock()
	h.Clients[userID] = conn
}

func (h *Hub) Unregister(userID uuid.UUID) {
	h.Mu.Lock()
	defer h.Mu.Unlock()
	delete(h.Clients, userID)
}
