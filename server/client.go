package server

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"os"
)

// Client represent service cmd client.
type Client struct {
}

// RunQuery handles stdin-stdout interactions for search queries.
func (c Client) RunQuery() {
	in := make(chan string)
	cancel := make(chan struct{})

	defer func() {
		close(cancel)
		close(in)
	}()

	go writeStdout(crawl(in, cancel))

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

func writeStdout(from <-chan *RepoResult) {
	for res := range from {
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
