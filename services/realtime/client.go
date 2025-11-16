package realtime

import (
	"encoding/json"
	"layer-api/types"
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	hub       *Hub
	conn      *websocket.Conn
	send      chan []byte
	userID    int
	noteID    int
	noteStore types.NoteStore
}

func NewClient(hub *Hub, conn *websocket.Conn, userID, noteID int, noteStore types.NoteStore) *Client {
	return &Client{
		hub:       hub,
		conn:      conn,
		send:      make(chan []byte, 256),
		userID:    userID,
		noteID:    noteID,
		noteStore: noteStore,
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		_ = c.conn.Close()
	}()

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("ws read error:", err)
			break
		}

		var msg types.RealtimeClientMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			c.sendError("invalid message format")
			continue
		}

		if msg.NoteID != 0 && msg.NoteID != c.noteID {
			c.sendError("note id mismatch")
			continue
		}

		switch msg.Type {
		case types.RealtimeMessageTypePatch:
			if err := c.noteStore.UpdateNoteContent(c.noteID, msg.Patch); err != nil {
				c.sendError("failed to save note")
				continue
			}

			serverMsg := types.RealtimeServerMessage{
				Type:    types.RealtimeMessageTypePatch,
				NoteID:  c.noteID,
				Patch:   msg.Patch,
				UserID:  c.userID,
				Version: msg.Version,
			}
			encoded, err := json.Marshal(serverMsg)
			if err != nil {
				continue
			}
			c.hub.Broadcast(c.noteID, encoded)

		default:
			c.sendError("unsupported message type")
		}
	}
}

func (c *Client) writePump() {
	defer func() {
		_ = c.conn.Close()
	}()

	for msg := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Println("ws write error:", err)
			break
		}
	}
}

func (c *Client) sendError(message string) {
	serverMsg := types.RealtimeServerMessage{
		Type:   types.RealtimeMessageTypeError,
		NoteID: c.noteID,
		Error:  message,
	}
	data, err := json.Marshal(serverMsg)
	if err != nil {
		return
	}
	select {
	case c.send <- data:
	default:
	}
}
