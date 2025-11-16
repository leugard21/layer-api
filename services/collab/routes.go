package collab

import (
	"layer-api/types"

	"github.com/gorilla/mux"
)

type Handler struct {
	store types.CollaboratorStore
}

func NewHandler(store types.CollaboratorStore) *Handler {
	return &Handler{store: store}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {

}
