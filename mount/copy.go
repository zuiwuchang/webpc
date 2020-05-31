package mount

import (
	"context"
	"errors"
	"fmt"
	"io"
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

// Copy 複製檔案
func Copy(ws *websocket.Conn, dir, srcDir string, timeout time.Duration) (names []string, e error) {
	c := &_CopyWorker{
		ws:     ws,
		ch:     make(chan _CopyMessage),
		cancel: make(chan struct{}),
	}
	e = c.done(dir, srcDir, timeout)
	c.close()
	names = c.Names
	return
}

type _CopyMessage struct {
	Cmd   int      `json:"cmd,omitempty"`
	Names []string `json:"names,omitempty"`
}
type _CopyWorker struct {
	ws      *websocket.Conn
	ch      chan _CopyMessage
	cancel  chan struct{}
	closed  int32
	Names   []string
	style   int
	srcRoot string
	dstRoot string
}

func (c *_CopyWorker) close() {
	if atomic.CompareAndSwapInt32(&c.closed, 0, 1) {
		c.ws.Close()
		close(c.cancel)
	}
}
func (c *_CopyWorker) read() {
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
		var msg _CopyMessage
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
func (c *_CopyWorker) getMessage(timeout time.Duration) (msg _CopyMessage, e error) {
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
func (c *_CopyWorker) waitMessage() (msg _CopyMessage, e error) {
	select {
	case <-c.cancel:
		e = context.Canceled
	case msg = <-c.ch:
	}
	return
}
func (c *_CopyWorker) writeExist(val string) error {
	return helper.WSWriteJSON(c.ws, gin.H{
		`cmd`: CmdExist,
		`val`: val,
	})
}
func (c *_CopyWorker) writeError(e error) error {
	return helper.WSWriteJSON(c.ws, gin.H{
		`cmd`:   CmdError,
		`error`: e.Error(),
	})
}
func (c *_CopyWorker) writeDone() error {
	return helper.WSWriteJSON(c.ws, gin.H{
		`cmd`: CmdDone,
	})
}
func (c *_CopyWorker) writeProgress(val string) error {
	return helper.WSWriteJSON(c.ws, gin.H{
		`cmd`: CmdProgress,
		`val`: val,
	})
}
func (c *_CopyWorker) done(dir, srcDir string, timeout time.Duration) (e error) {
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
	e = c.copy(dir, srcDir)
	if e != nil {
		return
	}
	c.writeDone()
	return
}
func (c *_CopyWorker) copy(dir, srcDir string) (e error) {
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

	c.dstRoot = dir
	c.srcRoot = srcDir
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
		name := c.Names[i]
		l := c.srcRoot + name
		r := c.dstRoot + name
		if strings.HasPrefix(l, r) || strings.HasPrefix(r, l) {
			e = errors.New(`Paths include each other`)
			c.writeError(e)
			return
		}

		e = c.copyRoot(name)
		if e != nil {
			return
		}
	}
	return
}
func (c *_CopyWorker) copyRoot(name string) (e error) {
	stat, e := os.Stat(c.srcRoot + name)
	if e != nil {
		c.writeError(e)
		return
	}
	if stat.IsDir() {
		e = c.copyRootDir(name)
	} else {
		e = c.copyFile(name)
	}
	return
}
func (c *_CopyWorker) copyRootDir(name string) (e error) {
	root := c.srcRoot + name
	e = filepath.Walk(root, func(path string, stat os.FileInfo, err error) (e error) {
		if err != nil {
			e = err
			c.writeError(e)
			return
		}
		name := path[len(c.srcRoot):]

		if stat.IsDir() {
			e = c.copyDir(stat, name)
		} else {
			e = c.copyFile(name)
		}
		return
	})
	if e != nil {
		c.writeError(e)
	}
	return
}
func (c *_CopyWorker) copyDir(stat os.FileInfo, name string) (e error) {
	e = c.writeProgress(name)
	if e != nil {
		return
	}
	e = os.MkdirAll(c.dstRoot+name, stat.Mode())
	if e != nil {
		c.writeError(e)
		return
	}
	return
}
func (c *_CopyWorker) copyFile(name string) (e error) {
	e = c.writeProgress(name)
	if e != nil {
		return
	}
	r, e := os.Open(c.srcRoot + name)
	if e != nil {
		c.writeError(e)
		return
	}
	defer r.Close()
	stat, e := r.Stat()
	if e != nil {
		c.writeError(e)
		return
	}
	if stat.IsDir() {
		e = c.copyDir(stat, name)
		return
	}
	w, e := c.createFile(c.dstRoot, name, stat.Mode())
	if e != nil {
		c.writeError(e)
		return
	}
	if w == nil {
		return
	}
	_, e = io.Copy(w, r)
	w.Close()
	if e != nil {
		c.writeError(e)
	}
	return
}
func (c *_CopyWorker) createFile(dir, name string, mode os.FileMode) (f *os.File, e error) {
	filename := filepath.Clean(dir + name)
	if c.style == CmdYesAll {
		f, e = os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode)
		return
	}
	f, e = os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_EXCL, mode)
	if !os.IsExist(e) {
		return
	}
	if c.style == CmdSkipAll {
		e = nil
		return
	}

	e0 := c.writeExist(name)
	if e0 != nil {
		return
	}
	msg, e0 := c.waitMessage()
	if e0 != nil {
		return
	}
	if msg.Cmd == CmdNo {
		e = context.Canceled
		return
	} else if msg.Cmd == CmdYes {
		f, e = os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode)
	} else if msg.Cmd == CmdYesAll {
		c.style = CmdYesAll
		f, e = os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode)
	} else if msg.Cmd == CmdSkip {
		e = nil
		return
	} else if msg.Cmd == CmdSkipAll {
		c.style = CmdSkipAll
		e = nil
		return
	} else {
		e = fmt.Errorf(`Expect yes or no,Unexpected instruction %v`, msg.Cmd)
		return
	}
	return
}
