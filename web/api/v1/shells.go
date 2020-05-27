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

	r.GET(``, h.CheckSession, h.list)
	r.GET(`:id/:cols/:rows`, h.CheckShell, h.connect)
	r.PATCH(`:id/name`, h.CheckShell, h.rename)
	r.DELETE(`:id`, h.CheckShell, h.remove)
}
func (h Shells) list(c *gin.Context) {
	session := h.BindSession(c)
	if session == nil {
		if ce := logger.Logger.Check(zap.ErrorLevel, c.FullPath()); ce != nil {
			ce.Write(
				zap.String(`method`, c.Request.Method),
				zap.String(`error`, `session nil`),
			)
		}
		return
	}
	manager := shell.Single()
	arrs := manager.List(session.Name)
	h.NegotiateData(c, http.StatusOK, arrs)
}
func (h Shells) connect(c *gin.Context) {
	session := h.BindSession(c)
	if session == nil {
		if ce := logger.Logger.Check(zap.ErrorLevel, c.FullPath()); ce != nil {
			ce.Write(
				zap.String(`method`, c.Request.Method),
				zap.String(`error`, `session nil`),
			)
		}
		return
	}
	var obj struct {
		ID   int64  `uri:"id"`
		Cols uint16 `uri:"cols"  binding:"required"`
		Rows uint16 `uri:"rows" binding:"required"`
	}
	e := h.BindURI(c, &obj)
	if e != nil {
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
		shell.WriteJSON(ws, gin.H{
			`cmd`:   shell.CmdError,
			`error`: e.Error(),
		})
		ws.Close()
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
			if msg.Cmd == shell.CmdResize {
				s.SetSize(msg.Cols, msg.Rows)
			} else if msg.Cmd == shell.CmdFontsize {
				s.SetFontsize(msg.Val)
			} else {
				if ce := logger.Logger.Check(zap.WarnLevel, c.FullPath()); ce != nil {
					ce.Write(
						zap.String("error", "not support websocket msg type"),
						zap.String(`method`, c.Request.Method),
						zap.String(`session`, session.String()),
						zap.String(`client ip`, c.ClientIP()),
						zap.Bool(`new`, newshell),
						zap.Int64(`id`, shellid),
						zap.Int(`cmd`, msg.Cmd),
					)
				}
			}
		} else {
			if ce := logger.Logger.Check(zap.WarnLevel, c.FullPath()); ce != nil {
				ce.Write(
					zap.String("error", "not support websocket type"),
					zap.String(`method`, c.Request.Method),
					zap.String(`session`, session.String()),
					zap.String(`client ip`, c.ClientIP()),
					zap.Bool(`new`, newshell),
					zap.Int64(`id`, shellid),
				)
			}
		}
	}
}
func (h Shells) rename(c *gin.Context) {
	session := h.BindSession(c)
	if session == nil {
		if ce := logger.Logger.Check(zap.ErrorLevel, c.FullPath()); ce != nil {
			ce.Write(
				zap.String(`method`, c.Request.Method),
				zap.String(`error`, `session nil`),
			)
		}
		return
	}
	var objURI struct {
		ID int64 `uri:"id" binding:"required"`
	}
	e := h.BindURI(c, &objURI)
	if e != nil {
		return
	}
	var obj struct {
		Name string `uri:"name" binding:"required"`
	}
	e = h.Bind(c, &obj)
	if e != nil {
		return
	}

	manager := shell.Single()
	e = manager.Rename(session.Name, objURI.ID, obj.Name)
	if e != nil {
		h.NegotiateError(c, http.StatusNotFound, e)
		return
	}
	c.Status(http.StatusNoContent)
	if ce := logger.Logger.Check(zap.InfoLevel, c.FullPath()); ce != nil {
		ce.Write(
			zap.String(`method`, c.Request.Method),
			zap.Int64(`id`, objURI.ID),
			zap.String(`val`, obj.Name),
		)
	}
}
func (h Shells) remove(c *gin.Context) {
	session := h.BindSession(c)
	if session == nil {
		if ce := logger.Logger.Check(zap.ErrorLevel, c.FullPath()); ce != nil {
			ce.Write(
				zap.String(`method`, c.Request.Method),
				zap.String(`error`, `session nil`),
			)
		}
		return
	}
	var objURI struct {
		ID int64 `uri:"id" binding:"required"`
	}
	e := h.BindURI(c, &objURI)
	if e != nil {
		return
	}
	manager := shell.Single()
	e = manager.Kill(session.Name, objURI.ID)
	if e != nil {
		return
	}
	if ce := logger.Logger.Check(zap.InfoLevel, c.FullPath()); ce != nil {
		ce.Write(
			zap.String(`method`, c.Request.Method),
			zap.Int64(`id`, objURI.ID),
		)
	}
}
