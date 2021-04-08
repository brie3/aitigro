package server

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"os"
	"time"
)

// Client represent service cmd client.
type Client struct {
}

// RunQuery handles stdin-stdout interactions for search queries.
func (c Client) RunQuery() {
	in := make(chan string)
	cancel := make(chan struct{})
	out := make(chan string)

	defer func() {
		close(in)
		close(out)
		close(cancel)
	}()

	go readStdin(in)

	resChan := crawl(out, cancel)

	for i := range in {
		out <- i
		time.Sleep(delay)

		res := <-resChan
		switch res.Error {
		case nil:
			pretty, err := json.MarshalIndent(res, ",", " ")
			if err != nil {
				log.Printf(decodeErrFormat, err)
				return
			}
			if _, err = io.WriteString(os.Stdout, string(pretty)); err != nil {
				log.Printf(osStdoutErrFormat, err)
				return
			}
		default:
			if _, err := io.WriteString(os.Stdout, res.Error.Error()); err != nil {
				log.Printf(osStdoutErrFormat, err)
				return
			}
		}
	}
}

func readStdin(out chan string) {
	var (
		err    error
		text   string
		reader = bufio.NewReader(os.Stdin)
	)
	for {
		text, err = reader.ReadString('\n')
		if err != nil {
			log.Printf(userReadErrFormat, err)
			return
		}
		select {
		case out <- text:
			if _, err = io.WriteString(os.Stdout, searchStartMessage); err != nil {
				log.Printf(osStdoutErrFormat, err)
				return
			}
		default:
			if _, err = io.WriteString(os.Stdout, busyString); err != nil {
				log.Printf(osStdoutErrFormat, err)
				return
			}
		}
	}
}
