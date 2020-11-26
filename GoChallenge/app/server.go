package app

import (
	"GoChallenge/message"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"
)

const (
	pingPeriod = time.Second * 10
)

var (
	busyMessage = message.Message{
		Type: message.MTInfo,
		Data: "Идёт поиск",
	}
	startMessage = message.Message{
		Type: message.MTInfo,
		Data: "Начинаю поиск",
	}
	pingMessage = message.Message{
		Type: message.MTPing,
	}
)

type Server struct {
	router   *chi.Mux
	upgrader *websocket.Upgrader
	done     chan bool
}

func (s *Server) Start() {
	s.router = chi.NewRouter()
	s.upgrader = &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	s.setUpHandlers()

	go func() {
		if err := http.ListenAndServe(":8888", s.router); err != nil {
			log.Println("Failed to start server: ", err)
		}
	}()
}

func (s *Server) setUpHandlers() {
	s.router.Handle("/*", http.FileServer(http.Dir("./web")))
	s.router.Get("/socket", s.socketHandler)
}

func (s *Server) socketHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf(upgradeErrFormat, err)
		return
	}

	pingTicker := time.NewTicker(pingPeriod)

	defer func() {
		pingTicker.Stop()
		ws.Close()
	}()

	var (
		block bool
		out   chan *RepoResult
		done  = make(chan struct{})
	)

	delay := func() {
		close(done)
	}

	for {
		var read message.Message
		if err = ws.ReadJSON(&read); err != nil {
			log.Printf(readMessageErrFormat, err)
			return
		}
		if read.Type == message.MTPong {
			continue
		}
		if !block {
			if err := ws.WriteJSON(&startMessage); err != nil {
				log.Printf(writeMessageErrFormat, err)
				return
			}
			go func() {
				if s, ok := read.Data.(string); ok {
					out = worker(s)
				}
			}()
			time.AfterFunc(timeout, delay)
			block = true
		} else if err = ws.WriteJSON(busyMessage); err != nil {
			log.Printf(writeMessageErrFormat, err)
			return
		}
		select {
		case <-pingTicker.C:
			if err = ws.WriteJSON(pingMessage); err != nil {
				log.Printf(writeMessageErrFormat, err)
				return
			}
		case <-done:
			select {
			case v := <-out:
				block = false
				done = make(chan struct{})
				if v.Error != nil {
					if err = ws.WriteJSON(message.Message{Type: message.MTError, Data: v.Error}); err != nil {
						log.Printf(writeMessageErrFormat, err)
						return
					}
				} else {
					if err = ws.WriteJSON(message.Message{Type: message.MTMessage, Data: v}); err != nil {
						log.Printf(writeMessageErrFormat, err)
						return
					}
				}
			}
		}
	}
}
