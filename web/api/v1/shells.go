package v1

import (
	"net/http"
	"os"
	"os/exec"

	"github.com/creack/pty"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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
	r.GET(`:id/:cols/:rows`, h.connect)
}
func (h Shells) list(c *gin.Context) {

}
func (h Shells) connect(c *gin.Context) {
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

	_, e = upgrader.Upgrade(c.Writer, c.Request, nil)
	if e != nil {
		return
	}

	cmd := exec.Command("/bin/bash", "-l")
	cmd.Env = append(os.Environ(), "TERM=xterm")
	tty, e := pty.StartWithSize(cmd, &pty.Winsize{
		Cols: obj.Cols,
		Rows: obj.Rows,
	})
}
