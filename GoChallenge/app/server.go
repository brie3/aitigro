package app

import (
	"GoChallenge/message"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"
)

var budyMessage = message.Message{
	Type: message.MTMessage,
	Data: "Идёт поиск",
}

var startMessage = message.Message{
	Type: message.MTMessage,
	Data: "Начинаю поиск",
}

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
		log.Println(err)
		return
	}
	defer ws.Close()

	var block bool

	for {
		msg := message.Message{}
		if err := ws.ReadJSON(&msg); err != nil {
			log.Println(err)
		}
		if msg.Type == message.MTPong {
			continue
		}
		if msg.Type == message.MTMessage {
			if !block {
				if err := ws.WriteJSON(startMessage); err != nil {
					log.Println(err)
				}
				block = true
				time.AfterFunc(timeout, func() {
					out := <-httpWorker(msg.Data.(string))
					if err := ws.WriteJSON(message.Message{
						Type: message.MTMessage,
						Data: out,
					}); err != nil {
						log.Println(err)
					}
					block = false
				})
			} else {
				if err := ws.WriteJSON(budyMessage); err != nil {
					log.Println(err)
				}
			}
		}
	}
}

func httpWorker(in string) chan *RepoResult {
	var (
		err    error
		resp   *http.Response
		result *RepoResult
	)
	out := make(chan *RepoResult)
	go func() {
		resp, err = http.Get(searchQuery + url.QueryEscape(golang+in))
		if err != nil {
			log.Printf("request error: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("search query failed: %s", resp.Status)
		}

		if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
			log.Printf("can't decode: %v", err)
		}
		out <- result
	}()
	return out
}
