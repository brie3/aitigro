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

	defer func() {
		close(in)
		close(cancel)
	}()

	go writeStdout(in, cancel)

	reader := bufio.NewReader(os.Stdin)
	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Printf(userReadErrFormat, err)
			return
		}
		select {
		case in <- text:
			if _, err := io.WriteString(os.Stdout, searchStartMessage); err != nil {
				log.Printf(osStdoutErrFormat, err)
				return
			}
		default:
			if _, err := io.WriteString(os.Stdout, busyString); err != nil {
				log.Printf(osStdoutErrFormat, err)
				return
			}
		}
	}
}

func writeStdout(in <-chan string, cancel <-chan struct{}) {
	out := make(chan string)
	defer close(out)

	var res *RepoResult
	resChan := crawl(out, cancel)
	for s := range in {
		out <- s
		time.Sleep(delay)

		res = <-resChan
		if res == nil {
			continue
		}

		switch res.Error {
		case nil:
			pretty, err := json.MarshalIndent(res, "", "	")
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
