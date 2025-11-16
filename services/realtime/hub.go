package realtime

import (
	"encoding/json"
	"layer-api/types"
)

type BroadcastMessage struct {
	NoteID int
	Data   []byte
}

type Hub struct {
	register   chan *Client
	unregister chan *Client
	broadcast  chan BroadcastMessage
	rooms      map[int]map[*Client]bool
}

func NewHub() *Hub {
	return &Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan BroadcastMessage),
		rooms:      make(map[int]map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			if h.rooms[c.noteID] == nil {
				h.rooms[c.noteID] = make(map[*Client]bool)
			}
			h.rooms[c.noteID][c] = true
			h.broadcastPresence(c.noteID)

		case c := <-h.unregister:
			clients, ok := h.rooms[c.noteID]
			if !ok {
				continue
			}
			if _, exists := clients[c]; exists {
				delete(clients, c)
				close(c.send)
				if len(clients) == 0 {
					delete(h.rooms, c.noteID)
				}
			}
			h.broadcastPresence(c.noteID)

		case msg := <-h.broadcast:
			clients, ok := h.rooms[msg.NoteID]
			if !ok {
				continue
			}
			for c := range clients {
				select {
				case c.send <- msg.Data:
				default:
					delete(clients, c)
					close(c.send)
					if len(clients) == 0 {
						delete(h.rooms, msg.NoteID)
					}
				}
			}
		}
	}
}

func (h *Hub) Broadcast(noteID int, data []byte) {
	h.broadcast <- BroadcastMessage{
		NoteID: noteID,
		Data:   data,
	}
}

func (h *Hub) broadcastPresence(noteID int) {
	clients, ok := h.rooms[noteID]
	if !ok {
		return
	}

	msg := types.RealtimeServerMessage{
		Type:       types.RealtimeMessageTypePresence,
		NoteID:     noteID,
		ActiveUser: len(clients),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	for c := range clients {
		select {
		case c.send <- data:
		default:
			delete(clients, c)
			close(c.send)
			if len(clients) == 0 {
				delete(h.rooms, noteID)
			}
		}
	}
}
