package helper

import (
	"github.com/gorilla/websocket"
)

// WSWriteJSON .
func WSWriteJSON(ws *websocket.Conn, obj interface{}) (e error) {
	b, e := Marshal(obj)
	if e != nil {
		return
	}
	e = ws.WriteMessage(websocket.TextMessage, b)
	return
}
