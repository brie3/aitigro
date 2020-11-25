package message

// Type represent ws message type.
type Type string

const (
	MTPing    Type = "ping"
	MTPong    Type = "pong"
	MTMessage Type = "message"
)

type Message struct {
	Type Type        `json:"type"`
	Data interface{} `json:"data,omitempty"`
}
