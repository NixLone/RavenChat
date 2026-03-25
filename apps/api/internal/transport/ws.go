package transport

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Hub struct {
	mu      sync.RWMutex
	clients map[uuid.UUID]map[*websocket.Conn]struct{}
}

func NewHub() *Hub {
	return &Hub{clients: map[uuid.UUID]map[*websocket.Conn]struct{}{}}
}

func (h *Hub) Handle(w http.ResponseWriter, r *http.Request, chatID uuid.UUID) {
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	h.mu.Lock()
	if _, ok := h.clients[chatID]; !ok {
		h.clients[chatID] = map[*websocket.Conn]struct{}{}
	}
	h.clients[chatID][conn] = struct{}{}
	h.mu.Unlock()
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			h.mu.Lock()
			delete(h.clients[chatID], conn)
			h.mu.Unlock()
			_ = conn.Close()
			return
		}
	}
}

func (h *Hub) BroadcastChat(chatID uuid.UUID, event any) {
	payload, _ := json.Marshal(event)
	h.mu.RLock()
	conns := h.clients[chatID]
	h.mu.RUnlock()
	for c := range conns {
		_ = c.WriteMessage(websocket.TextMessage, payload)
	}
}
