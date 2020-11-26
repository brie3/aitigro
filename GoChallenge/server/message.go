package server

import (
	"time"
)

const (
	MTPing    Type = "ping"
	MTPong    Type = "pong"
	MTError   Type = "error"
	MTMessage Type = "message"
	MTInfo    Type = "info"

	pingPeriod = time.Second * 10
	delay      = time.Second * 10

	internalErrorFormat   = "internal error"
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

// Type represent websocket message type.
type Type string

// Message represent websocket message.
type Message struct {
	Type Type        `json:"type"`
	Data interface{} `json:"data,omitempty"`
}

// RepoResult represent GitHub search result.
type RepoResult struct {
	Error error   `json:"-"`
	Total int     `json:"total_count"`
	Repos []*Repo `json:"items"`
}

// Repo represent GitHub repo basic struct.
type Repo struct {
	HTMLURL     string    `json:"html_url"`
	Title       string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

var (
	busyMessage = Message{
		Type: MTInfo,
		Data: "Идёт поиск",
	}
	startMessage = Message{
		Type: MTInfo,
		Data: "Начинаю поиск",
	}
	pingMessage = Message{
		Type: MTPing,
	}
	errorMessage = Message{
		Type: MTError,
		Data: internalErrorFormat,
	}
)
