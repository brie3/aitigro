package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

func crawl(in <-chan string, cancel <-chan struct{}) <-chan *RepoResult {
	out := make(chan *RepoResult)
	go func() {
		defer close(out)
		client := http.Client{Timeout: delay}
		for i := range in {
			resp, err := client.Get(searchQuery + url.QueryEscape(filter+i))
			time.Sleep(delay)
			if err != nil {
				out <- &RepoResult{Error: err}
				return
			}

			if resp.StatusCode != http.StatusOK {
				resp.Body.Close()
				out <- &RepoResult{Error: fmt.Errorf(badStatusCode, resp.StatusCode, i)}
				return
			}

			var result RepoResult
			if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
				resp.Body.Close()
				out <- &RepoResult{Error: err}
				return
			}
			resp.Body.Close()
			select {
			case <-cancel:
				return
			case out <- &result:
			}
		}
	}()
	return out
}
