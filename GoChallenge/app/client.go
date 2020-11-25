package app

import (
	"bufio"
	"encoding/json"
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
	Total int     `json:"total_count"`
	Repos []*Repo `json:"items"`
}

// Issue struct for the GitHub issue tracker.
type Repo struct {
	HTMLURL   string    `json:"html_url"`
	Title     string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

const (
	busyString  = "Идёт поиск"
	searchQuery = "https://api.github.com/search/repositories?q="
	golang      = "golang"
)

type Client struct {
}

func (c Client) RunQuery() {
	in := make(chan string)

	go reader(in)

	var block bool
	for {
		select {
		case v := <-in:
			if !block {
				block = true
				time.AfterFunc(timeout, func() {
					out := <-worker(v)
					_, err := io.WriteString(os.Stdout, out)
					if err != nil {
						log.Printf("can't write string into os.stdout: %v", err)
					}
					block = false
				})
			} else {
				log.Println("Идёт поиск")
			}
		}
	}
}

func reader(out chan string) {
	for {
		text, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			log.Printf("can't read message from user: %v", err)
		} else {
			out <- text
		}
	}
}

func worker(in string) chan string {
	var (
		err    error
		resp   *http.Response
		result *RepoResult
		pretty []byte
	)
	out := make(chan string)
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

		pretty, err = json.MarshalIndent(result, "", "	")
		if err != nil {
			log.Printf("can't decode: %v", err)
		}
		out <- string(pretty)
	}()
	return out
}
