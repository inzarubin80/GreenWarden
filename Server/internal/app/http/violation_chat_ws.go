package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/inzarubin80/Server/internal/app/uhttp"
	wsapp "github.com/inzarubin80/Server/internal/app/ws"
	"github.com/inzarubin80/Server/internal/model"
)

type (
	violationChatService interface {
		PostViolationChatMessage(ctx context.Context, userID model.UserID, violationID model.ViolationID, text string, isSystem bool) (*model.ViolationChatMessage, error)
	}

	// ViolationChatWSHandler handles WebSocket connections for violation chat.
	ViolationChatWSHandler struct {
		name    string
		service violationChatService
		hub     *wsapp.Hub
	}
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// NewViolationChatWSHandler creates a new WS handler for violation chat.
func NewViolationChatWSHandler(name string, svc violationChatService, hub *wsapp.Hub) *ViolationChatWSHandler {
	return &ViolationChatWSHandler{name: name, service: svc, hub: hub}
}

// wsIncomingMessage describes messages coming from client.
type wsIncomingMessage struct {
	Type        string `json:"type"`         // "subscribe" | "message"
	ViolationID string `json:"violation_id"` // for both types
	Text        string `json:"text"`         // for "message"
}

func (h *ViolationChatWSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Пока что простой вариант: требуем user_id как query-параметр (например, для отладки),
	// позже можно заменить на реальную аутентификацию через сессии/токены.
	q := r.URL.Query()
	userIDStr := q.Get("user_id")
	if userIDStr == "" {
		uhttp.SendErrorResponse(w, http.StatusUnauthorized, "user_id is required")
		return
	}
	userIDNum, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil || userIDNum <= 0 {
		uhttp.SendErrorResponse(w, http.StatusUnauthorized, "invalid user_id")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := &wsapp.Client{
		Conn:       conn,
		UserID:     model.UserID(userIDNum),
		Violations: make(map[model.ViolationID]struct{}),
		Send:       make(chan any, 32),
		Hub:        h.hub,
	}

	h.hub.RegisterClient(client)

	go client.WritePump()
	client.ReadPump(func(raw []byte) {
		var msg wsIncomingMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			return
		}

		switch msg.Type {
		case "subscribe":
			if msg.ViolationID != "" {
				client.Subscribe(model.ViolationID(msg.ViolationID))
			}
		case "message":
			if msg.ViolationID == "" || msg.Text == "" {
				return
			}
			_, err := h.service.PostViolationChatMessage(
				r.Context(),
				client.UserID,
				model.ViolationID(msg.ViolationID),
				msg.Text,
				false,
			)
			if err != nil {
				return
			}
			// New message events будут доставлены всем подписчикам через Hub.
		}
	})
}
