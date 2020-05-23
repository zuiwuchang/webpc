package v1

import (
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gitlab.com/king011/webpc/logger"
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
	session := h.BindSession(c)
	if session == nil {
		return
	}
	manager := shell.Single()
	arrs := manager.List(session.Name)
	h.NegotiateData(c, http.StatusOK, arrs)
}
func (h Shells) connect(c *gin.Context) {
	session := h.BindSession(c)
	if session == nil {
		return
	}
	var obj struct {
		ID   int64  `uri:"id"`
		Cols uint16 `uri:"cols"  binding:"required"`
		Rows uint16 `uri:"rows" binding:"required"`
	}
	e := c.ShouldBindUri(&obj)
	if e != nil {
		h.NegotiateError(c, http.StatusBadRequest, e)
		return
	}
	var newshell bool
	var shellid int64
	if obj.ID == 0 {
		newshell = true
		shellid = time.Now().UTC().Unix()
	} else {
		shellid = obj.ID
	}

	ws, e := upgrader.Upgrade(c.Writer, c.Request, nil)
	if e != nil {
		return
	}
	manager := shell.Single()
	s, e := manager.Attach(ws, session.Name, shellid, obj.Cols, obj.Rows, newshell)
	if e != nil {
		ws.Close()
		if ce := logger.Logger.Check(zap.WarnLevel, c.FullPath()); ce != nil {
			ce.Write(
				zap.Error(e),
				zap.String(`method`, c.Request.Method),
				zap.String(`session`, session.String()),
				zap.String(`client ip`, c.ClientIP()),
				zap.Bool(`new`, newshell),
				zap.Int64(`id`, shellid),
			)
		}
		shell.WriteMessage(ws, shell.DataTypeError, e.Error())
		return
	}
	if ce := logger.Logger.Check(zap.InfoLevel, c.FullPath()); ce != nil {
		ce.Write(
			zap.Error(e),
			zap.String(`method`, c.Request.Method),
			zap.String(`session`, session.String()),
			zap.String(`client ip`, c.ClientIP()),
			zap.Bool(`new`, newshell),
			zap.Int64(`id`, shellid),
		)
	}
	defer func() {
		ws.Close()
		s.Unattack(ws)
	}()

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
				if ce := logger.Logger.Check(zap.ErrorLevel, c.FullPath()); ce != nil {
					ce.Write(
						zap.Error(e),
						zap.String(`method`, c.Request.Method),
						zap.String(`session`, session.String()),
						zap.String(`client ip`, c.ClientIP()),
						zap.Bool(`new`, newshell),
						zap.Int64(`id`, shellid),
					)
				}
				continue
			}
			if msg.What == shell.DataTypeResize {
				e = s.SetSize(msg.Cols, msg.Rows)
			}
		}
	}
}
