package server

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

func (s *Server) socketHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf(upgradeErrFormat, err)
		return
	}

	in := make(chan string)
	cancel := make(chan struct{})

	defer func() {
		ws.Close()
		close(cancel)
		close(in)
	}()

	go ticker(ws)
	go writeWS(ws, crawl(in, cancel))

	for {
		var read Message
		if err = ws.ReadJSON(&read); err != nil {
			log.Printf(readMessageErrFormat, err)
			return
		}
		if read.Type == MTPong {
			continue
		}

		query, ok := read.Data.(string)
		switch ok {
		case true:
		default:
			if err := ws.WriteJSON(&badMessage); err != nil {
				log.Printf(writeMessageErrFormat, err)
				return
			}
		}

		select {
		case in <- query:
			if err := ws.WriteJSON(&startMessage); err != nil {
				log.Printf(writeMessageErrFormat, err)
				return
			}
		default:
			if err = ws.WriteJSON(busyMessage); err != nil {
				log.Printf(writeMessageErrFormat, err)
				return
			}
		}
	}
}

func writeWS(ws *websocket.Conn, from <-chan *RepoResult) {
	for res := range from {
		switch res.Error {
		case nil:
			if err := ws.WriteJSON(Message{Type: MTMessage, Data: res}); err != nil {
				log.Printf(writeMessageErrFormat, err)
				return
			}
		default:
			log.Println(res.Error.Error())
			if err := ws.WriteJSON(errorMessage); err != nil {
				log.Printf(writeMessageErrFormat, err)
				return
			}
		}
	}
}

func ticker(ws *websocket.Conn) {
	pingTicker := time.NewTicker(pingPeriod)
	defer pingTicker.Stop()

	for range pingTicker.C {
		if err := ws.WriteJSON(pingMessage); err != nil {
			log.Printf(writeMessageErrFormat, err)
			return
		}
	}
}
