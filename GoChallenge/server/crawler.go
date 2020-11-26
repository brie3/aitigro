package server

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
)

func crawl(in string, cancel chan struct{}) chan *RepoResult {
	out := make(chan *RepoResult)
	go func() {
		defer close(out)
		resp, err := http.Get(searchQuery + url.QueryEscape(filter+in))
		if err != nil {
			log.Println(requestQueryErrFormat, err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Println(responseErrFormat, resp.Status)
			return
		}
		var result RepoResult
		if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
			log.Println(decodeErrFormat, err)
			return
		}
		select {
		case <-cancel:
			return
		case out <- &result:
		}
	}()
	return out
}
