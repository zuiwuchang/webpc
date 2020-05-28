package shell

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gitlab.com/king011/webpc/helper"
)

// Message .
type Message struct {
	Cmd  int    `json:"cmd,omitempty"`
	Cols uint16 `json:"cols,omitempty"`
	Rows uint16 `json:"rows,omitempty"`
	Val  int    `json:"val,omitempty"`
}

// Unmarshal .
func (m *Message) Unmarshal(data []byte) error {
	return helper.Unmarshal(data, m)
}

// WriteInfo .
func WriteInfo(ws *websocket.Conn, id int64, name string, started int64, fontSize int) (e error) {
	m := gin.H{
		`cmd`:     CmdInfo,
		`id`:      id,
		`name`:    name,
		`started`: started,
	}
	if fontSize >= 5 {
		m[`fontSize`] = fontSize
	}
	b, e := helper.Marshal(m)
	if e != nil {
		return
	}
	e = ws.WriteMessage(websocket.TextMessage, b)
	return
}
