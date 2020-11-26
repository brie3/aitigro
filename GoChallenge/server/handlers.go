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
	defer ws.Close()

	out := make(chan *Message)
	defer close(out)

	go ticker(ws)
	go writeWs(ws, out)

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

func writeWs(ws *websocket.Conn, in <-chan *Message) {
	cancel := make(chan struct{})
	defer close(cancel)

	var err error
	for i := range in {
		s, ok := i.Data.(string)
		if !ok {
			continue
		}
		out := crawl(s, cancel)
		time.Sleep(delay)
		for j := range out {
			if j.Error == nil {
				if err = ws.WriteJSON(Message{Type: MTMessage, Data: j}); err != nil {
					log.Printf(osStdoutErrFormat, err)
					return
				}
			} else if err = ws.WriteJSON(errorMessage); err != nil {
				log.Printf(osStdoutErrFormat, err)
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
