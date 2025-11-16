package collab

import (
	"database/sql"
	"errors"
	"fmt"
	"layer-api/types"
	"layer-api/utils"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type Handler struct {
	collabStore types.CollaboratorStore
	noteStore   types.NoteStore
}

func NewHandler(collabStore types.CollaboratorStore, noteStore types.NoteStore) *Handler {
	return &Handler{
		collabStore: collabStore,
		noteStore:   noteStore,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.Handle("/notes/{id}/collaborators", utils.AuthMiddleware(http.HandlerFunc(h.HandleAddCollaborator))).Methods("POST")
	router.Handle("/notes/{id}/collaborators", utils.AuthMiddleware(http.HandlerFunc(h.HandleListCollaborators))).Methods("GET")
	router.Handle("/notes/{id}/collaborators/{userId}", utils.AuthMiddleware(http.HandlerFunc(h.HandleRemoveCollaborator))).Methods("DELETE")
}

func (h *Handler) HandleAddCollaborator(w http.ResponseWriter, r *http.Request) {
	userID, ok := utils.GetUserIDFromContext(r.Context())
	if !ok || userID <= 0 {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid token"))
		return
	}

	noteId, err := parseIDFromVars(r, "id")
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	note, err := h.noteStore.GetNoteByID(noteId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.WriteError(w, http.StatusNotFound, fmt.Errorf("note not found"))
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	if note.OwnerID != userID {
		utils.WriteError(w, http.StatusForbidden, fmt.Errorf("only owner can manage collaborators"))
		return
	}

	var payload types.AddCollaboratorPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}
	if err := utils.Validate.Struct(payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	canEdit := true
	if payload.CanEdit != nil {
		canEdit = *payload.CanEdit
	}

	if err := h.collabStore.AddCollaborator(noteId, payload.UserID, canEdit); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"message": "collaborator added/updated",
	})
}

func (h *Handler) HandleListCollaborators(w http.ResponseWriter, r *http.Request) {
	userID, ok := utils.GetUserIDFromContext(r.Context())
	if !ok || userID <= 0 {
		utils.WriteError(w, http.StatusUnauthorized, errors.New("invalid token"))
		return
	}

	noteID, err := parseIDFromVars(r, "id")
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	note, err := h.noteStore.GetNoteByID(noteID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.WriteError(w, http.StatusNotFound, errors.New("note not found"))
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	if note.OwnerID != userID {
		utils.WriteError(w, http.StatusForbidden, errors.New("only owner can view collaborators"))
		return
	}

	collabs, err := h.collabStore.ListCollaborators(noteID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, collabs)
}

func (h *Handler) HandleRemoveCollaborator(w http.ResponseWriter, r *http.Request) {
	userID, ok := utils.GetUserIDFromContext(r.Context())
	if !ok || userID <= 0 {
		utils.WriteError(w, http.StatusUnauthorized, errors.New("invalid token"))
		return
	}

	noteID, err := parseIDFromVars(r, "id")
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	note, err := h.noteStore.GetNoteByID(noteID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.WriteError(w, http.StatusNotFound, errors.New("note not found"))
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	if note.OwnerID != userID {
		utils.WriteError(w, http.StatusForbidden, errors.New("only owner can remove collaborators"))
		return
	}

	targetUserID, err := parseIDFromVars(r, "userId")
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.collabStore.RemoveCollaborator(noteID, targetUserID); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"message": "collaborator removed",
	})
}

func parseIDFromVars(r *http.Request, key string) (int, error) {
	vars := mux.Vars(r)
	raw, ok := vars[key]
	if !ok || raw == "" {
		return 0, fmt.Errorf("missing id")
	}

	id, err := strconv.Atoi(raw)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid id")
	}

	return id, nil
}
