package app

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

const timeout = 10 * time.Second

// IssuesSearchResult struct for the GitHub search result.
type RepoResult struct {
	Error error   `json:"-"`
	Total int     `json:"total_count"`
	Repos []*Repo `json:"items"`
}

// Issue struct for the GitHub issue tracker.
type Repo struct {
	HTMLURL     string    `json:"html_url"`
	Title       string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

const (
	busyString            = "Идёт поиск\n"
	searchQuery           = "https://api.github.com/search/repositories?q="
	filter                = "golang "
	osStdoutErrFormat     = "can't write string into os.stdout: %v"
	writeMessageErrFormat = "can't write message into socket: %v"
	readMessageErrFormat  = "can't read message from socket: %v"
	userReadErrFormat     = "can't read message from user: %v"
	upgradeErrFormat      = "can't upgrade connection: %v"
	requestQueryErrFormat = "request error: %v"
	responseErrFormat     = "search query failed: %s"
	decodeErrFormat       = "can't decode: %v"
)

type Client struct {
}

func (c Client) RunQuery() {
	var (
		block  bool
		err    error
		pretty []byte
		done   = make(chan struct{})
		out    chan *RepoResult
	)

	in := make(chan string)
	go reader(in)

	defer close(done)

	delay := func() {
		close(done)
	}

	for {
		select {
		case v := <-in:
			if !block {
				go func() {
					out = worker(v)
				}()
				time.AfterFunc(timeout, delay)
				block = true
			} else if _, err = io.WriteString(os.Stdout, busyString); err != nil {
				log.Printf(osStdoutErrFormat, err)
			}
		case <-done:
			for v := range out {
				block = false
				done = make(chan struct{})
				if v.Error == nil {
					pretty, err = json.MarshalIndent(v, "", "	")
					if err != nil {
						log.Printf(decodeErrFormat, err)
					}
					_, err = io.WriteString(os.Stdout, string(pretty))
					if err != nil {
						log.Printf(osStdoutErrFormat, err)
					}
				} else if _, err = io.WriteString(os.Stdout, v.Error.Error()); err != nil {
					log.Printf(osStdoutErrFormat, err)
				}
			}
		}
	}
}

func reader(out chan string) {
	var (
		err    error
		text   string
		reader = bufio.NewReader(os.Stdin)
	)
	for {
		text, err = reader.ReadString('\n')
		if err != nil {
			log.Printf(userReadErrFormat, err)
		} else {
			out <- text
		}
	}
}

func worker(in string) chan *RepoResult {
	out := make(chan *RepoResult)
	go func() {
		var tmp RepoResult
		defer func() {
			out <- &tmp
		}()

		resp, err := http.Get(searchQuery + url.QueryEscape(filter+in))
		if err != nil {
			tmp.Error = fmt.Errorf(requestQueryErrFormat, err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			tmp.Error = fmt.Errorf(responseErrFormat, resp.Status)
			return
		}

		if err = json.NewDecoder(resp.Body).Decode(&tmp); err != nil {
			tmp.Error = fmt.Errorf(decodeErrFormat, err)
			return
		}
	}()
	return out
}
