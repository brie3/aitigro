package server

import (
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"
)

// Server handles http interactions.
type Server struct {
	router   *chi.Mux
	upgrader *websocket.Upgrader
	// done     chan bool
}

// Start starts server.
func (s *Server) Start() {
	s.router = chi.NewRouter()
	s.upgrader = &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	s.setupHandlers()

	port := ":__PORT__"

	go func() {
		if err := http.ListenAndServe(port, s.router); err != nil {
			log.Println("Failed to start server: ", err)
		}
	}()
}

func (s *Server) setupHandlers() {
	s.router.Handle("/*", http.FileServer(http.Dir("./web")))
	s.router.Get("/socket", s.socketHandler)
}
