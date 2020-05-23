package v1

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gitlab.com/king011/webpc/shell"
	"gitlab.com/king011/webpc/web"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Shells .
type Shells struct {
	web.Helper
}

// Register impl IHelper
func (h Shells) Register(router *gin.RouterGroup) {
	r := router.Group(`/shells`)

	r.GET(``, h.list)
	r.GET(`:id/:cols/:rows`, h.CheckShell, h.connect)
}
func (h Shells) list(c *gin.Context) {

}
func (h Shells) connect(c *gin.Context) {
	session := h.BindSession(c)
	if session == nil {
		return
	}
	var obj struct {
		ID   string `uri:"id"  binding:"required"`
		Cols uint16 `uri:"cols"  binding:"required"`
		Rows uint16 `uri:"rows" binding:"required"`
	}
	e := c.ShouldBindUri(&obj)
	if e != nil {
		h.NegotiateError(c, http.StatusBadRequest, e)
		return
	}
	u, e := uuid.NewUUID()
	if e != nil {
		h.NegotiateError(c, http.StatusBadRequest, e)
		return
	}
	shellid := u.String()

	ws, e := upgrader.Upgrade(c.Writer, c.Request, nil)
	if e != nil {
		return
	}
	defer ws.Close()

	s, e := shell.New(`/bin/bash`, `-l`)
	if e != nil {
		shell.WriteMessage(ws, shell.DataTypeError, e.Error())
		return
	}

	// 運行 shell
	e = s.Run(ws, session.Name, shellid, obj.Cols, obj.Rows)
	if e != nil {
		shell.WriteMessage(ws, shell.DataTypeError, e.Error())
		return
	}

	// 讀取 websocket
	var msg shell.Message
	for {
		t, p, e := ws.ReadMessage()
		if e != nil {
			break
		}
		if t == websocket.BinaryMessage {
			s.Write(p)
		} else if t == websocket.TextMessage {
			e = msg.Unmarshal(p)
			if e != nil {
				continue
			}
			if msg.What == shell.DataTypeResize {
				e = s.SetSize(msg.Cols, msg.Rows)
			}
		}
	}
}
