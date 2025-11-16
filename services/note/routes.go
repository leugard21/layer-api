package note

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
	store types.NoteStore
}

func NewHandler(store types.NoteStore) *Handler {
	return &Handler{store: store}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.Handle("/notes", utils.AuthMiddleware(http.HandlerFunc(h.handleCreateNote))).Methods("POST")
	router.Handle("/notes", utils.AuthMiddleware(http.HandlerFunc(h.handleListNotes))).Methods("GET")
	router.Handle("/notes/{id}", utils.AuthMiddleware(http.HandlerFunc(h.handleGetNote))).Methods("GET")
	router.Handle("/notes/{id}", utils.AuthMiddleware(http.HandlerFunc(h.handleUpdateNote))).Methods("PATCH")
	router.Handle("/notes/{id}/archive", utils.AuthMiddleware(http.HandlerFunc(h.handleArchiveNote))).Methods("POST")
}

func (h *Handler) handleCreateNote(w http.ResponseWriter, r *http.Request) {
	userID, ok := utils.GetUserIDFromContext(r.Context())
	if !ok || userID <= 0 {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid token"))
		return
	}

	var payload types.CreateNotePayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}
	if err := utils.Validate.Struct(payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	n := types.Note{
		OwnerID: userID,
		Title:   payload.Title,
		Content: payload.Content,
	}

	id, err := h.store.CreateNote(n)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	created, err := h.store.GetNoteByID(id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, created)
}

func (h *Handler) handleListNotes(w http.ResponseWriter, r *http.Request) {
	userID, ok := utils.GetUserIDFromContext(r.Context())
	if !ok || userID <= 0 {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid token"))
		return
	}

	notes, err := h.store.ListNotesByOwner(userID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, notes)
}

func (h *Handler) handleGetNote(w http.ResponseWriter, r *http.Request) {
	userID, ok := utils.GetUserIDFromContext(r.Context())
	if !ok || userID <= 0 {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid token"))
		return
	}

	id, err := parseIDFromVars(r)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	n, err := h.store.GetNoteByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.WriteError(w, http.StatusNotFound, errors.New("note not found"))
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	if n.OwnerID != userID {
		utils.WriteError(w, http.StatusNotFound, fmt.Errorf("note not found"))
		return
	}

	utils.WriteJSON(w, http.StatusOK, n)
}

func (h *Handler) handleUpdateNote(w http.ResponseWriter, r *http.Request) {
	userID, ok := utils.GetUserIDFromContext(r.Context())
	if !ok || userID <= 0 {
		utils.WriteError(w, http.StatusUnauthorized, fmt.Errorf("invalid token"))
		return
	}

	id, err := parseIDFromVars(r)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	existing, err := h.store.GetNoteByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.WriteError(w, http.StatusNotFound, fmt.Errorf("note not found"))
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	if existing.OwnerID != userID {
		utils.WriteError(w, http.StatusNotFound, errors.New("note not found"))
		return
	}

	var payload types.UpdateNotePayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}
	if err := utils.Validate.Struct(payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if payload.Title == nil && payload.Content == nil {
		utils.WriteError(w, http.StatusBadRequest, errors.New("no fields to update"))
		return
	}

	updated := types.Note{
		ID:      existing.ID,
		OwnerID: existing.OwnerID,
		Title:   existing.Title,
		Content: existing.Content,
	}

	if payload.Title != nil {
		updated.Title = *payload.Title
	}
	if payload.Content != nil {
		updated.Content = *payload.Content
	}

	if err := h.store.UpdateNote(updated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.WriteError(w, http.StatusNotFound, errors.New("note not found"))
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	n, err := h.store.GetNoteByID(id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, n)
}

func (h *Handler) handleArchiveNote(w http.ResponseWriter, r *http.Request) {
	userID, ok := utils.GetUserIDFromContext(r.Context())
	if !ok || userID <= 0 {
		utils.WriteError(w, http.StatusUnauthorized, errors.New("invalid token"))
		return
	}

	id, err := parseIDFromVars(r)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.store.ArchiveNote(id, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.WriteError(w, http.StatusNotFound, errors.New("note not found"))
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "note archived",
	})
}

func parseIDFromVars(r *http.Request) (int, error) {
	vars := mux.Vars(r)
	rawID, ok := vars["id"]
	if !ok || rawID == "" {
		return 0, fmt.Errorf("missing note id")
	}

	id, err := strconv.Atoi(rawID)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid note id")
	}

	return id, nil
}
