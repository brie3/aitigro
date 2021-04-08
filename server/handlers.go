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
	out := make(chan *Message)
	cancel := make(chan struct{})
	defer func() {
		ws.Close()
		close(out)
		close(cancel)
	}()

	go ticker(ws)
	go writeWs(ws, out, cancel)

	for {
		var read Message
		if err = ws.ReadJSON(&read); err != nil {
			log.Printf(readMessageErrFormat, err)
			return
		}
		if read.Type == MTPong {
			continue
		}
		select {
		case out <- &read:
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

func writeWs(ws *websocket.Conn, in <-chan *Message, cancel <-chan struct{}) {
	out := make(chan string)
	defer close(out)

	resChan := crawl(out, cancel)
	for i := range in {
		s, ok := i.Data.(string)
		if !ok {
			if err := ws.WriteJSON(&badMessage); err != nil {
				log.Printf(writeMessageErrFormat, err)
			}
			continue
		}
		out <- s
		time.Sleep(delay)
		res := <-resChan
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
