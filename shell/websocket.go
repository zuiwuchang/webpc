package shell

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Message .
type Message struct {
	Cmd  int    `json:"cmd,omitempty"`
	Cols uint16 `json:"cols,omitempty"`
	Rows uint16 `json:"rows,omitempty"`
}

// Unmarshal .
func (m *Message) Unmarshal(data []byte) error {
	return json.Unmarshal(data, m)
}

// WriteJSON .
func WriteJSON(ws *websocket.Conn, obj interface{}) (e error) {
	b, e := json.Marshal(obj)
	if e != nil {
		return
	}
	e = ws.WriteMessage(websocket.TextMessage, b)
	return
}

// WriteInfo .
func WriteInfo(ws *websocket.Conn, id int64, name string) (e error) {
	b, e := json.Marshal(gin.H{
		`cmd`:  CmdInfo,
		`id`:   id,
		`name`: name,
	})
	if e != nil {
		return
	}
	e = ws.WriteMessage(websocket.TextMessage, b)
	return
}
