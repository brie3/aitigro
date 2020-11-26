package message

// Type represent ws message type.
type Type string

const (
	MTPing    Type = "ping"
	MTPong    Type = "pong"
	MTError   Type = "error"
	MTMessage Type = "message"
	MTInfo    Type = "info"
)

type Message struct {
	Type Type        `json:"type"`
	Data interface{} `json:"data,omitempty"`
}
