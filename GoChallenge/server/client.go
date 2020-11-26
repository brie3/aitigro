package server

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"os"
	"time"
)

type Client struct {
}

func (c Client) RunQuery() {
	in := make(chan string)
	defer close(in)

	go readStdin(in)

	for i := range in {
		out := make(chan *RepoResult)
		go crawl(i, out)
		time.Sleep(delay)
		for j := range out {
			if j.Error == nil {
				pretty, err := json.MarshalIndent(j, "", "	")
				if err != nil {
					log.Printf(decodeErrFormat, err)
				}
				if _, err = io.WriteString(os.Stdout, string(pretty)); err != nil {
					log.Printf(osStdoutErrFormat, err)
					return
				}
			} else if _, err := io.WriteString(os.Stdout, j.Error.Error()); err != nil {
				log.Printf(osStdoutErrFormat, err)
				return
			}
			close(out)
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
		if err == nil {
			select {
			case out <- text:
			default:
				if _, err = io.WriteString(os.Stdout, busyString); err != nil {
					log.Printf(osStdoutErrFormat, err)
					return
				}
			}
		} else {
			log.Printf(userReadErrFormat, err)
			return
		}
	}
}
