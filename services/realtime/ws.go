package realtime

import (
	"database/sql"
	"errors"
	"fmt"
	"layer-api/services/collab"
	"layer-api/services/note"
	"layer-api/utils"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Handler struct {
	hub         *Hub
	noteStore   *note.Store
	collabStore *collab.Store
}

func NewHandler(hub *Hub, noteStore *note.Store, collabStore *collab.Store) *Handler {
	return &Handler{
		hub:         hub,
		noteStore:   noteStore,
		collabStore: collabStore,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.Handle("/ws/notes/{id}", http.HandlerFunc(h.handleNoteWS)).Methods("GET")
}

func (h *Handler) handleNoteWS(w http.ResponseWriter, r *http.Request) {
	userID, ok := utils.GetUserIDFromContext(r.Context())
	if !ok || userID <= 0 {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid token"))
		return
	}

	vars := mux.Vars(r)
	rawID := vars["id"]
	noteID, err := strconv.Atoi(rawID)
	if err != nil || noteID <= 0 {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid note id"))
		return
	}

	n, err := h.noteStore.GetNoteByID(noteID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.WriteError(w, http.StatusNotFound, fmt.Errorf("note not found"))
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	if n.OwnerID != userID {
		okCollab, err := h.collabStore.IsCollaborator(noteID, userID)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}
		if !okCollab {
			utils.WriteError(w, http.StatusForbidden, fmt.Errorf("no access to this note"))
			return
		}
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	_ = conn
	_ = userID
	_ = noteID
	_ = h.hub
}
