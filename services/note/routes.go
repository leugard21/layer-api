package note

import (
	"layer-api/types"

	"github.com/gorilla/mux"
)

type Handler struct {
	store types.NoteStore
}

func NewHandler(store types.NoteStore) *Handler {
	return &Handler{store: store}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {

}
