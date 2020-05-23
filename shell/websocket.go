package shell

import (
	"github.com/gorilla/websocket"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Message .
type Message struct {
	What    int    `json:"what,omitempty"`
	Message string `json:"msg,omitempty"`
	Cols    uint16 `json:"cols,omitempty"`
	Rows    uint16 `json:"rows,omitempty"`
}

// Unmarshal .
func (m *Message) Unmarshal(data []byte) error {
	return json.Unmarshal(data, m)
}

// WriteMessage .
func WriteMessage(ws *websocket.Conn, what int, msg string) (e error) {
	b, e := json.Marshal(Message{
		What:    what,
		Message: msg,
	})
	if e != nil {
		return
	}

	e = ws.WriteMessage(websocket.TextMessage, b)
	return
}
