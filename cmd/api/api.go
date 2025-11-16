package api

import (
	"database/sql"
	"layer-api/services/collab"
	"layer-api/services/note"
	"layer-api/services/user"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type APIServer struct {
	addr string
	db   *sql.DB
}

func NewAPIServer(addr string, db *sql.DB) *APIServer {
	return &APIServer{
		addr: addr,
		db:   db,
	}
}

func (s *APIServer) Run() error {
	router := mux.NewRouter()
	subrouter := router.PathPrefix("/api/v1").Subrouter()

	userStore := user.NewStore(s.db)
	userHandler := user.NewHandler(userStore)
	userHandler.RegisterRoutes(subrouter)

	noteStore := note.NewStore(s.db)
	noteHandler := note.NewHandler(noteStore)
	noteHandler.RegisterRoutes(subrouter)

	collabStore := collab.NewStore(s.db)
	collabHandler := collab.NewHandler(collabStore, noteStore)
	collabHandler.RegisterRoutes(subrouter)

	log.Println("Listening on", s.addr)

	return http.ListenAndServe(s.addr, router)
}
