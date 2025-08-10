package api

import (
	"CodeStream/src/resources"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Consider restricting this in production
	},
}

type Client struct {
	ID       string
	Username string
	Conn     *websocket.Conn
	Hub      *Hub
	Send     chan []byte // Buffered channel for outbound messages
	mu       sync.Mutex  // Protects writes to websocket
}

type Hub struct {
	Clients     map[*Client]bool
	SessionID   string
	Interview   *resources.Interview
	mu          sync.RWMutex
	interviewMu sync.Mutex

	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte

	shutdown chan struct{}
	done     chan struct{}
}

type Message struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data,omitempty"`
}

var (
	Sessions   = make(map[string]*Hub)
	sessionsMu sync.RWMutex // Protects Sessions map
)

// Constants for timeouts and buffer sizes
const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
	sendBufferSize = 256
)

func NewHub(sessionID string, cache *resources.Cache) (*Hub, error) {
	interview, err, _ := resources.GetOrCreateInterviewSession(cache, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to create interview session: %w", err)
	}

	return &Hub{
		SessionID:  sessionID,
		Clients:    make(map[*Client]bool),
		Interview:  &interview,
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
		shutdown:   make(chan struct{}),
		done:       make(chan struct{}),
	}, nil
}

func GetHub(sessionID string, cache *resources.Cache) (*Hub, error) {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()

	if hub, exists := Sessions[sessionID]; exists {
		return hub, nil
	}

	// Create new hub
	hub, err := NewHub(sessionID, cache)
	if err != nil {
		return nil, err
	}

	Sessions[sessionID] = hub
	go hub.run() // Start the hub goroutine

	return hub, nil
}

func (h *Hub) run() {
	defer close(h.done)

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.Clients[client] = true
			clientCount := len(h.Clients)
			h.mu.Unlock()

			log.Printf("Client %s joined session %s. Total clients: %d",
				client.Username, h.SessionID, clientCount)

			// Notify others about new user
			msg := Message{
				Type: "user_joined",
				Data: map[string]interface{}{
					"username": client.Username,
				},
			}
			msgBytes, err := json.Marshal(msg)
			if err != nil {
				log.Printf("Error marshalling message: %v", err)
				return
			}
			h.broadcastToOthers(client, msgBytes)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
			}
			clientCount := len(h.Clients)
			h.mu.Unlock()

			log.Printf("Client %s left session %s. Remaining clients: %d",
				client.Username, h.SessionID, clientCount)

			msg := Message{
				Type: "user_left",
				Data: map[string]interface{}{
					"username": client.Username,
				},
			}
			msgBytes, err := json.Marshal(msg)
			if err != nil {
				log.Printf("Error marshalling message: %v", err)
				return
			}
			h.broadcastToOthers(client, msgBytes)

			// Clean up empty sessions
			if clientCount == 0 {
				sessionsMu.Lock()
				delete(Sessions, h.SessionID)
				sessionsMu.Unlock()
				log.Printf("Session %s cleaned up", h.SessionID)
				return // Exit hub goroutine
			}

		case message := <-h.broadcast:
			h.mu.Lock()
			clientsToRemove := make([]*Client, 0)

			for client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					clientsToRemove = append(clientsToRemove, client)
				}
			}

			for _, client := range clientsToRemove {
				delete(h.Clients, client)
				close(client.Send)
				log.Printf("Removed unresponsive client: %s", client.Username)
			}
			h.mu.Unlock()

		case <-h.shutdown:
			h.mu.Lock()
			for client := range h.Clients {
				close(client.Send)
				_ = client.Conn.Close()
			}
			h.Clients = make(map[*Client]bool)
			h.mu.Unlock()
			return
		}
	}
}

func (h *Hub) broadcastToOthers(sender *Client, msg []byte) {

	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.Clients {
		if client != sender {
			select {
			case client.Send <- msg:
				// Message queued successfully
			default:
				// Channel full, client will be cleaned up in next broadcast cycle
				log.Printf("Client %s channel full, will be cleaned up", client.Username)
			}
		}
	}
}

func (h *Hub) Shutdown() {
	close(h.shutdown)
	// Wait for hub to finish
	select {
	case <-h.done:
	case <-time.After(5 * time.Second):
		log.Printf("Hub shutdown timeout for session %s", h.SessionID)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Use client mutex to protect websocket writes
			c.mu.Lock()
			err := c.Conn.WriteMessage(websocket.TextMessage, message)
			c.mu.Unlock()

			if err != nil {
				log.Printf("Error writing message to client %s: %v", c.Username, err)
				return
			}

		case <-ticker.C:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))

			c.mu.Lock()
			err := c.Conn.WriteMessage(websocket.PingMessage, nil)
			c.mu.Unlock()

			if err != nil {
				log.Printf("Error sending ping to client %s: %v", c.Username, err)
				return
			}
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		_ = c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		var msg Message
		err := c.Conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error for client %s: %v", c.Username, err)
			}
			break
		}

		// Process different message types
		switch msg.Type {
		case "code_patch":
			if err := c.processCodePatch(msg); err != nil {
				log.Printf("Error processing code patch from %s: %v", c.Username, err)
				errorMsg := Message{
					Type: "error",
					Data: map[string]interface{}{
						"message": err.Error(),
						"type":    "code_patch_error",
					},
				}
				if jsonData, marshalErr := json.Marshal(errorMsg); marshalErr == nil {
					select {
					case c.Send <- jsonData:
					default:
					}
				}
			}
		case "cursor_position":
			c.processCursorPosition(msg)

		case "edit_lang":
			c.processEditLang(msg)

		case "refresh":
			c.sendCurrentState()

		default:
			log.Printf("Unknown message type from %s: %s", c.Username, msg.Type)
		}
	}
}

func (c *Client) sendCurrentState() {
	// Send initial session data with current code state
	c.Hub.interviewMu.Lock()
	currentCode, patches, version, err := c.Hub.Interview.GetCurrentCode()
	c.Hub.interviewMu.Unlock()

	users := []map[string]string{}
	for user, _ := range c.Hub.Clients {
		users = append(users, map[string]string{
			"username": user.Username,
			"id":       user.ID,
		})
	}

	initialData := Message{
		Type: "session_init",
		Data: map[string]interface{}{
			"session_id":   c.Hub.SessionID,
			"user_id":      c.ID,
			"current_code": currentCode,
			"lang":         c.Hub.Interview.Language,
			"version":      version,
			"patches":      patches,
			"users":        users,
		},
	}

	if err != nil {
		initialData.Data = map[string]interface{}{
			"session_id": c.Hub.SessionID,
			"user_id":    c.ID,
			"error":      "Failed to load current code state",
		}
	}

	if jsonData, err := json.Marshal(initialData); err == nil {
		select {
		case c.Send <- jsonData:
		case <-time.After(5 * time.Second):
			log.Printf("Timeout sending initial data to client %s", c.Username)
		}
	}
}

func (c *Client) processCodePatch(msg Message) error {
	dataBytes, err := json.Marshal(msg.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal patch data: %w", err)
	}

	var patch resources.CodePatch
	if err := json.Unmarshal(dataBytes, &patch); err != nil {
		return fmt.Errorf("failed to unmarshal patch: %w", err)
	}

	if patch.Operation != "add" && patch.Operation != "remove" && patch.Operation != "replace" {
		return fmt.Errorf("invalid operation: %s", patch.Operation)
	}

	c.Hub.interviewMu.Lock()
	err = c.Hub.Interview.AddCodePatch(patch)
	c.Hub.interviewMu.Unlock()

	if err != nil {
		if strings.Contains(err.Error(), "version mismatch") {
			c.sendCurrentState()
			return nil
		}
		return fmt.Errorf("failed to add code patch: %w", err)
	}

	broadcastMsg := Message{
		Type: "code_patch",
		Data: map[string]interface{}{
			"username":  c.Username,
			"version":   patch.Version,
			"op":        patch.Operation,
			"start_pos": patch.StartPos,
			"end_pos":   patch.EndPos,
			"content":   patch.Content,
		},
	}

	msgBytes, _ := json.Marshal(broadcastMsg)
	c.Hub.broadcastToOthers(c, msgBytes)
	return nil
}

func (c *Client) processCursorPosition(msg Message) {
	position, ok := msg.Data["position"]
	if !ok {
		log.Printf("Missing position data in cursor_position message from %s", c.Username)
		return
	}

	broadcastMsg := Message{
		Type: "cursor_position",
		Data: map[string]interface{}{
			"username": c.Username,
			"position": position,
		},
	}

	msgBytes, _ := json.Marshal(broadcastMsg)

	c.Hub.broadcastToOthers(c, msgBytes)
}

func (c *Client) processEditLang(msg Message) {
	newLang, ok := msg.Data["lang"].(string)
	if !ok {
		log.Printf("Missing lang data in edit_lang message from %s", c.Username)
		errorMsg := Message{
			Type: "error",
			Data: map[string]interface{}{
				"message": "Missing lang data",
				"type":    "edit_lang_error",
			},
		}
		if jsonData, err := json.Marshal(errorMsg); err == nil {
			select {
			case c.Send <- jsonData:
			default:
			}
		}
		return
	}

	err := c.Hub.Interview.EditLanguage(newLang)
	if err != nil {
		log.Printf("Error editing language to %s: %v", newLang, err)
		errorMsg := Message{
			Type: "error",
			Data: map[string]interface{}{
				"message": err.Error(),
				"type":    "edit_lang_error",
			},
		}
		if jsonData, err := json.Marshal(errorMsg); err == nil {
			select {
			case c.Send <- jsonData:
			default:
			}
		}
		return
	}

	// Broadcast the language change to all clients
	broadcastMsg := Message{
		Type: "lang_change",
		Data: map[string]interface{}{
			"lang": newLang,
		},
	}
	msgBytes, _ := json.Marshal(broadcastMsg)
	c.Hub.broadcast <- msgBytes
}

func LiveStreamCoding(c *gin.Context) {
	sessionID := c.Query("session_id")
	username := c.Query("username")

	if sessionID == "" || username == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "session_id and username query parameters are required",
		})
		return
	}

	cache := resources.NewCacheContext()
	if !cache.Exists(fmt.Sprintf("session:%s:state", sessionID)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session does not exist"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	hub, err := GetHub(sessionID, cache)
	if err != nil {
		log.Printf("Failed to get hub for session %s: %v", sessionID, err)
		_ = conn.Close()
		return
	}

	client := &Client{
		Username: username,
		Conn:     conn,
		Hub:      hub,
		ID:       username + strconv.FormatInt(time.Now().UnixNano(), 10),
		Send:     make(chan []byte, sendBufferSize),
	}

	hub.register <- client
	go client.writePump()

	client.sendCurrentState()

	client.readPump()
}
