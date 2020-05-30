package mount

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gitlab.com/king011/webpc/helper"
	"gitlab.com/king011/webpc/logger"
	"go.uber.org/zap"
)

// Cut 剪下檔案
func Cut(ws *websocket.Conn, dir, srcDir string, timeout time.Duration) (names []string, e error) {
	c := &_CutWorker{
		ws:     ws,
		ch:     make(chan _CutMessage),
		cancel: make(chan struct{}),
	}
	e = c.done(dir, srcDir, timeout)
	c.close()
	names = c.Names
	return
}

type _CutMessage struct {
	Cmd   int      `json:"cmd,omitempty"`
	Names []string `json:"names,omitempty"`
}
type _CutWorker struct {
	ws     *websocket.Conn
	ch     chan _CutMessage
	cancel chan struct{}
	closed int32
	keys   map[string]bool
	Names  []string
}

func (c *_CutWorker) close() {
	if atomic.CompareAndSwapInt32(&c.closed, 0, 1) {
		c.ws.Close()
		close(c.cancel)
	}
}
func (c *_CutWorker) read() {
	for {
		t, b, e := c.ws.ReadMessage()
		if e != nil {
			break
		} else if t != websocket.TextMessage {
			if ce := logger.Logger.Check(zap.WarnLevel, `cut not support message type`); ce != nil {
				ce.Write(
					zap.Int("type", t),
				)
			}
			break
		}
		var msg _CutMessage
		e = helper.Unmarshal(b, &msg)
		if e != nil {
			if ce := logger.Logger.Check(zap.WarnLevel, `cut unmarshal message error`); ce != nil {
				ce.Write(
					zap.Error(e),
				)
			}
			break
		}
		if msg.Cmd == CmdHeart {
			continue
		}
		select {
		case <-c.cancel:
			c.close()
			return
		case c.ch <- msg:
		}
	}
	c.close()
}
func (c *_CutWorker) getMessage(timeout time.Duration) (msg _CutMessage, e error) {
	timer := time.NewTimer(timeout)
	select {
	case <-c.cancel:
		e = context.Canceled
		if !timer.Stop() {
			<-timer.C
		}
	case <-timer.C:
		e = context.DeadlineExceeded
	case msg = <-c.ch:
		if !timer.Stop() {
			<-timer.C
		}
	}
	return
}
func (c *_CutWorker) waitMessage() (msg _CutMessage, e error) {
	select {
	case <-c.cancel:
		e = context.Canceled
	case msg = <-c.ch:
	}
	return
}
func (c *_CutWorker) writeError(e error) error {
	return helper.WSWriteJSON(c.ws, gin.H{
		`cmd`:   CmdError,
		`error`: e.Error(),
	})
}
func (c *_CutWorker) writeDone() error {
	return helper.WSWriteJSON(c.ws, gin.H{
		`cmd`: CmdDone,
	})
}
func (c *_CutWorker) writeProgress(val string) error {
	return helper.WSWriteJSON(c.ws, gin.H{
		`cmd`: CmdProgress,
		`val`: val,
	})
}
func (c *_CutWorker) done(dir, srcDir string, timeout time.Duration) (e error) {
	go c.read()
	msg, e := c.getMessage(timeout)
	if e != nil {
		return
	}
	if msg.Cmd != CmdInit {
		e = fmt.Errorf(`Expect init,Unexpected instruction %v`, msg.Cmd)
		c.writeError(e)
		return
	}
	c.Names = msg.Names
	e = c.cut(dir, srcDir)
	if e != nil {
		return
	}
	c.writeDone()
	return
}
func (c *_CutWorker) cut(dir, srcDir string) (e error) {
	count := len(c.Names)
	if count == 0 {
		e = errors.New(`names nil`)
		c.writeError(e)
		return
	}

	if !strings.HasSuffix(dir, Separator) {
		dir += Separator
	}
	if !strings.HasSuffix(srcDir, Separator) {
		srcDir += Separator
	}

	for i := 0; i < count; i++ {
		str := filepath.Clean(dir + c.Names[i])
		if !strings.HasPrefix(str, dir) {
			e = fmt.Errorf(`name not support : %v`, c.Names[i])
			c.writeError(e)
			return
		}
		str = filepath.Base(str)
		if str != c.Names[i] || str == `.` || str == `..` {
			e = fmt.Errorf(`name not support : %v`, c.Names[i])
			c.writeError(e)
			return
		}
	}

	for i := 0; i < count; i++ {
		e = c.writeProgress(c.Names[i])
		if e != nil {
			return
		}
		oldpath := srcDir + c.Names[i]
		newpath := dir + c.Names[i]
		if oldpath == newpath {
			continue
		}
		e = os.Rename(oldpath, newpath)
		if e != nil {
			c.writeError(e)
			return
		}
	}
	return
}
