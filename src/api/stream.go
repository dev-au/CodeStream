package api

import (
	"CodeStream/src"
	"CodeStream/src/resources"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true // Consider restricting this in production
	},
}

type Client struct {
	Username string
	Conn     *websocket.Conn
	Hub      *Hub
	Send     chan []byte
	mu       sync.Mutex
}

type Hub struct {
	Clients     map[string]*Client
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
	sessionsMu sync.RWMutex
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 2048
	sendBufferSize = 2048
)

func NewHub(sessionID string, cache *resources.Cache) (*Hub, error) {
	interview, err := resources.GetInterviewSession(cache, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get interview session: %w", err)
	}

	return &Hub{
		SessionID:  sessionID,
		Clients:    make(map[string]*Client),
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

	hub, err := NewHub(sessionID, cache)
	if err != nil {
		return nil, err
	}

	Sessions[sessionID] = hub
	go hub.run()

	return hub, nil
}

func (h *Hub) run() {
	defer close(h.done)

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.Clients[client.Username] = client
			clientCount := len(h.Clients)
			h.mu.Unlock()

			log.Printf("Client %s joined session %s. Total clients: %d",
				client.Username, h.SessionID, clientCount)

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
			if _, ok := h.Clients[client.Username]; ok {
				delete(h.Clients, client.Username)
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
				return
			}

		case message := <-h.broadcast:
			h.mu.Lock()
			clientsToRemove := make([]*Client, 0)

			for _, client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					clientsToRemove = append(clientsToRemove, client)
				}
			}

			for _, client := range clientsToRemove {
				delete(h.Clients, client.Username)
				close(client.Send)
				log.Printf("Removed unresponsive client: %s", client.Username)
			}
			h.mu.Unlock()

		case <-h.shutdown:
			h.mu.Lock()
			for _, client := range h.Clients {
				close(client.Send)
				_ = client.Conn.Close()
			}
			h.Clients = make(map[string]*Client)
			h.mu.Unlock()
			return
		}
	}
}

func (h *Hub) broadcastToOthers(sender *Client, msg []byte) {

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.Clients {
		if client != sender {
			select {
			case client.Send <- msg:
			default:
				// Channel full, client will be cleaned up in next broadcast cycle
				log.Printf("Client %s channel full, will be cleaned up", client.Username)
			}
		}
	}
}

func (h *Hub) Shutdown() {
	close(h.shutdown)
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
		case "code_run":
			c.processRunCode()
		case "cursor_select":
			c.processCursorSelect(msg)
		case "edit_lang":
			c.processEditLang(msg)
		case "refresh":
			c.sendCurrentState()

		default:
			log.Printf("Unknown message type from %s: %s", c.Username, msg.Type)
		}
	}
}

func (c *Client) processRunCode() {
	if !c.Hub.Interview.CanRun() {
		errorMsg := Message{
			Type: "error",
			Data: map[string]interface{}{
				"message": fmt.Sprintf("Rate limit exceeded"),
			},
		}
		msgBytes, _ := json.Marshal(errorMsg)
		c.Send <- msgBytes
		return
	}

	c.Hub.interviewMu.Lock()
	currentCode := c.Hub.Interview.CompactCodePatches()
	c.Hub.interviewMu.Unlock()

	if currentCode == "" {
		msg := Message{
			Type: "code_res",
			Data: map[string]interface{}{
				"std_out":   "",
				"std_err":   "",
				"exit_code": -1,
				"error":     "empty code",
				"duration":  "0.00",
			},
		}
		msgBytes, _ := json.Marshal(msg)
		c.Send <- msgBytes
		return
	}

	req := resources.RunRequest{Language: c.Hub.Interview.Language, Code: currentCode}

	resp, err := resources.RunUserCode(c.Hub.Interview.Cache.Ctx, src.Config.CodeWorkDir, req)
	if err != nil {
		errorMsg := Message{
			Type: "error",
			Data: map[string]interface{}{
				"message": err.Error(),
			},
		}
		msgBytes, _ := json.Marshal(errorMsg)
		c.Send <- msgBytes
		return
	}

	msg := Message{
		Type: "code_res",
		Data: map[string]interface{}{
			"std_out":   resp.Stdout,
			"std_err":   resp.Stderr,
			"exit_code": resp.ExitCode,
			"error":     resp.Error,
			"duration":  resp.Duration,
		},
	}
	msgBytes, _ := json.Marshal(msg)

	c.Send <- msgBytes
	c.Hub.broadcastToOthers(c, msgBytes)

}

func (c *Client) sendCurrentState() {
	c.Hub.interviewMu.Lock()
	currentCode, patches, version, err := c.Hub.Interview.GetCurrentCode()
	c.Hub.interviewMu.Unlock()

	var users []map[string]string
	for username, _ := range c.Hub.Clients {
		users = append(users, map[string]string{
			"username": username,
		})
	}

	initialData := Message{
		Type: "session_init",
		Data: map[string]interface{}{
			"session_id":   c.Hub.SessionID,
			"current_code": currentCode,
			"lang":         c.Hub.Interview.Language,
			"version":      version,
			"patches":      patches,
			"users":        users,
			"username":     c.Username,
		},
	}

	if err != nil {
		initialData.Data = map[string]interface{}{
			"session_id": c.Hub.SessionID,
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

func (c *Client) processCursorSelect(msg Message) {
	startVal, ok := msg.Data["start_pos"]
	if !ok {
		log.Printf("Missing start_pos data in cursor_position message from %s", c.Username)
		return
	}
	startFloat, ok := startVal.(float64)
	if !ok {
		log.Printf("Invalid start_pos type in cursor_position message from %s", c.Username)
		return
	}
	startPos := int(startFloat)

	endVal, ok := msg.Data["end_pos"]
	if !ok {
		log.Printf("Missing end_pos data in cursor_position message from %s", c.Username)
		return
	}
	endFloat, ok := endVal.(float64)
	if !ok {
		log.Printf("Invalid end_pos type in cursor_position message from %s", c.Username)
		return
	}
	endPos := int(endFloat)
	broadcastMsg := Message{
		Type: "cursor_select",
		Data: map[string]interface{}{
			"username":  c.Username,
			"start_pos": startPos,
			"end_pos":   endPos,
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

	broadcastMsg := Message{
		Type: "edit_lang",
		Data: map[string]interface{}{
			"lang": newLang,
		},
	}
	msgBytes, _ := json.Marshal(broadcastMsg)
	c.Hub.broadcastToOthers(c, msgBytes)
}

func LiveStreamCoding(c *gin.Context) {
	sessionID := c.Query("session_id")

	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "session_id query parameters is required",
		})
		return
	}

	if hub, exists := Sessions[sessionID]; exists && len(hub.Clients) >= 300 {
		c.JSON(400, gin.H{"error": "Too many users"})
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

	ttl := cache.GetTTL(fmt.Sprintf("session:%s:state", sessionID))
	timer := time.AfterFunc(ttl, func() {
		_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Session expired"))
		_ = conn.Close()
	})

	defer timer.Stop()

	hub, err := GetHub(sessionID, cache)
	if err != nil {
		log.Printf("Failed to get hub for session %s: %v", sessionID, err)
		_ = conn.Close()
		return
	}
	hub.interviewMu.Lock()
	username := fmt.Sprintf("User%d", rand.Intn(360)+1)
	for {
		if _, exists := hub.Clients[username]; !exists {
			break
		}
		username = fmt.Sprintf("User%d", rand.Intn(360)+1)
	}
	client := &Client{
		Username: username,
		Conn:     conn,
		Hub:      hub,
		Send:     make(chan []byte, sendBufferSize),
	}
	hub.interviewMu.Unlock()

	hub.register <- client
	client.sendCurrentState()

	go client.writePump()
	client.readPump()

}
