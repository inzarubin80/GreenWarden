package ws

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/inzarubin80/Server/internal/model"
)

// Message is an internal hub message envelope sent to subscribers.
type Message struct {
	ViolationID model.ViolationID
	UserID      *model.UserID
	Payload     any
}

// Client represents a single WebSocket connection.
type Client struct {
	Conn       *websocket.Conn
	UserID     model.UserID
	Violations map[model.ViolationID]struct{}
	Send       chan any
	Hub        *Hub
	mu         sync.RWMutex
	closedOnce sync.Once
}

// Hub manages WebSocket clients grouped by violationID.
type Hub struct {
	clientsByViolation map[model.ViolationID]map[*Client]struct{}
	register           chan *Client
	unregister         chan *Client
	broadcast          chan *Message
	mu                 sync.RWMutex
}

// NewHub returns a new Hub instance.
func NewHub() *Hub {
	return &Hub{
		clientsByViolation: make(map[model.ViolationID]map[*Client]struct{}),
		register:           make(chan *Client),
		unregister:         make(chan *Client),
		broadcast:          make(chan *Message, 256),
	}
}

// Run starts the main event loop of the hub.
func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.addClient(c)
		case c := <-h.unregister:
			h.removeClient(c)
		case msg := <-h.broadcast:
			h.dispatch(msg)
		}
	}
}

// RegisterClient registers a client with the hub.
func (h *Hub) RegisterClient(c *Client) {
	h.register <- c
}

// UnregisterClient unregisters a client from the hub.
func (h *Hub) UnregisterClient(c *Client) {
	h.unregister <- c
}

// Subscribe adds the client to a violation room.
func (c *Client) Subscribe(violationID model.ViolationID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Violations == nil {
		c.Violations = make(map[model.ViolationID]struct{})
	}
	if _, ok := c.Violations[violationID]; ok {
		return
	}
	c.Violations[violationID] = struct{}{}

	c.Hub.mu.Lock()
	defer c.Hub.mu.Unlock()
	if _, ok := c.Hub.clientsByViolation[violationID]; !ok {
		c.Hub.clientsByViolation[violationID] = make(map[*Client]struct{})
	}
	c.Hub.clientsByViolation[violationID][c] = struct{}{}
}

// Unsubscribe removes the client from a violation room.
func (c *Client) Unsubscribe(violationID model.ViolationID) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.Violations, violationID)

	c.Hub.mu.Lock()
	defer c.Hub.mu.Unlock()
	if clients, ok := c.Hub.clientsByViolation[violationID]; ok {
		delete(clients, c)
		if len(clients) == 0 {
			delete(c.Hub.clientsByViolation, violationID)
		}
	}
}

// Close closes the client connection and unregisters it.
func (c *Client) Close() {
	c.closedOnce.Do(func() {
		c.Hub.UnregisterClient(c)
		close(c.Send)
		_ = c.Conn.Close()
	})
}

// ReadPump reads messages from the WebSocket connection and passes raw payloads
// to the provided callback for application-level handling.
func (c *Client) ReadPump(onMessage func(raw []byte)) {
	defer c.Close()
	c.Conn.SetReadLimit(64 * 1024)
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
		if onMessage != nil {
			onMessage(message)
		}
	}
}

// WritePump drains the send channel and writes JSON messages to the websocket.
func (c *Client) WritePump() {
	defer c.Close()
	for msg := range c.Send {
		if err := c.Conn.WriteJSON(msg); err != nil {
			break
		}
	}
}

func (h *Hub) addClient(c *Client) {
	// initial registration does not subscribe to any violation yet
}

func (h *Hub) removeClient(c *Client) {
	c.mu.Lock()
	defer c.mu.Unlock()

	h.mu.Lock()
	defer h.mu.Unlock()

	for vID := range c.Violations {
		if clients, ok := h.clientsByViolation[vID]; ok {
			delete(clients, c)
			if len(clients) == 0 {
				delete(h.clientsByViolation, vID)
			}
		}
	}
}

func (h *Hub) dispatch(msg *Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.clientsByViolation[msg.ViolationID]
	if !ok {
		return
	}

	for c := range clients {
		// if UserID is set, send only to that user
		if msg.UserID != nil && c.UserID != *msg.UserID {
			continue
		}
		select {
		case c.Send <- msg.Payload:
		default:
			// slow client, close
			go c.Close()
		}
	}
}

// BroadcastViolationMessage sends a payload to all clients subscribed to the violation.
func (h *Hub) BroadcastViolationMessage(violationID model.ViolationID, payload any) error {
	select {
	case h.broadcast <- &Message{ViolationID: violationID, Payload: payload}:
	default:
		log.Printf("ws hub broadcast channel full, dropping message for violation %s", violationID)
	}
	return nil
}

// SendViolationMessageToUser sends a payload to a specific user within the violation room.
func (h *Hub) SendViolationMessageToUser(violationID model.ViolationID, userID model.UserID, payload any) error {
	select {
	case h.broadcast <- &Message{ViolationID: violationID, UserID: &userID, Payload: payload}:
	default:
		log.Printf("ws hub broadcast channel full, dropping direct message for violation %s", violationID)
	}
	return nil
}
