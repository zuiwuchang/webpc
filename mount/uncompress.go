package mount

import (
	"archive/tar"
	"archive/zip"
	"compress/bzip2"
	"compress/gzip"
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

type _UncompressMessage struct {
	Cmd  int    `json:"cmd,omitempty"`
	Name string `json:"name,omitempty"`
}
type _UncompressWorker struct {
	Name   string
	ws     *websocket.Conn
	ch     chan _UncompressMessage
	cancel chan struct{}
	closed int32
	keys   map[string]bool
	Names  []string
	style  int
}

// Uncompress 執行 解壓
func Uncompress(ws *websocket.Conn, dir string, timeout time.Duration) (name string, e error) {
	c := &_UncompressWorker{
		ws:     ws,
		ch:     make(chan _UncompressMessage),
		cancel: make(chan struct{}),
	}
	e = c.done(dir, timeout)
	c.close()
	name = c.Name
	return
}
func (c *_UncompressWorker) close() {
	if atomic.CompareAndSwapInt32(&c.closed, 0, 1) {
		c.ws.Close()
		close(c.cancel)
	}
}
func (c *_UncompressWorker) read() {
	for {
		t, b, e := c.ws.ReadMessage()
		if e != nil {
			break
		} else if t != websocket.TextMessage {
			if ce := logger.Logger.Check(zap.WarnLevel, `uncompress not support message type`); ce != nil {
				ce.Write(
					zap.Int("type", t),
				)
			}
			break
		}
		var msg _UncompressMessage
		e = helper.Unmarshal(b, &msg)
		if e != nil {
			if ce := logger.Logger.Check(zap.WarnLevel, `uncompress unmarshal message error`); ce != nil {
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
func (c *_UncompressWorker) getMessage(timeout time.Duration) (msg _UncompressMessage, e error) {
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
func (c *_UncompressWorker) waitMessage() (msg _UncompressMessage, e error) {
	select {
	case <-c.cancel:
		e = context.Canceled
	case msg = <-c.ch:
	}
	return
}
func (c *_UncompressWorker) done(dir string, timeout time.Duration) (e error) {
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
	c.Name = msg.Name
	e = c.uncompress(dir)
	if e != nil {
		return
	}
	c.writeDone()
	return
}
func (c *_UncompressWorker) writeError(e error) error {
	return helper.WSWriteJSON(c.ws, gin.H{
		`cmd`:   CmdError,
		`error`: e.Error(),
	})
}
func (c *_UncompressWorker) writeDone() error {
	return helper.WSWriteJSON(c.ws, gin.H{
		`cmd`: CmdDone,
	})
}
func (c *_UncompressWorker) writeProgress(val string) error {
	return helper.WSWriteJSON(c.ws, gin.H{
		`cmd`: CmdProgress,
		`val`: val,
	})
}
func (c *_UncompressWorker) writeExist(val string) error {
	return helper.WSWriteJSON(c.ws, gin.H{
		`cmd`: CmdExist,
		`val`: val,
	})
}
func (c *_UncompressWorker) uncompress(dir string) (e error) {
	if !strings.HasSuffix(dir, Separator) {
		dir += Separator
	}
	dst := filepath.Clean(dir + c.Name)
	if !strings.HasPrefix(dst, dir) {
		e = fmt.Errorf(`name not support : %v`, c.Name)
		c.writeError(e)
		return
	}
	str := filepath.Base(dst)
	if str != c.Name || str == `.` || str == `..` {
		e = fmt.Errorf(`name not support : %v`, c.Name)
		c.writeError(e)
		return
	}
	f, e := os.Open(dst)
	if e != nil {
		c.writeError(e)
		return
	}
	e = c.uncompressFile(f, dir, strings.ToLower(c.Name))
	f.Close()
	return
}
func (c *_UncompressWorker) uncompressFile(f *os.File, dir, ext string) (e error) {
	if strings.HasSuffix(ext, `.tar.gz`) {
		e = c.uncompressTarGZ(f, dir)
	} else if strings.HasSuffix(ext, `.tar.bz2`) {
		e = c.uncompressTarBZ2(f, dir)
	} else if strings.HasSuffix(ext, `.tar`) {
		e = c.uncompressTar(tar.NewReader(f), dir)
	} else if strings.HasSuffix(ext, `.zip`) {
		e = c.uncompressZip(f, dir)
	} else {
		e = errors.New(`not support uncompress`)
		c.writeError(e)
	}
	return
}
func (c *_UncompressWorker) uncompressTarGZ(f *os.File, dir string) (e error) {
	gf, e := gzip.NewReader(f)
	if e != nil {
		c.writeError(e)
		return
	}
	e = c.uncompressTar(tar.NewReader(gf), dir)
	gf.Close()
	return
}
func (c *_UncompressWorker) uncompressTarBZ2(f *os.File, dir string) (e error) {
	return c.uncompressTar(tar.NewReader(bzip2.NewReader(f)), dir)
}
func (c *_UncompressWorker) uncompressTar(r *tar.Reader, dir string) (e error) {
	for {
		var header *tar.Header
		header, e = r.Next()
		if e != nil {
			if e == io.EOF {
				e = nil
			}
			if e != nil {
				c.writeError(e)
			}
			break
		}
		switch header.Typeflag {
		case tar.TypeDir:
			// 更新進度
			e = c.writeProgress(header.Name)
			if e != nil {
				return
			}
			e = c.uncompressDir(dir, header.Name, os.FileMode(header.Mode))
			if e != nil {
				c.writeError(e)
				return
			}
		case tar.TypeReg:
			// 更新進度
			e = c.writeProgress(header.Name)
			if e != nil {
				return
			}

			e = c.doneUncompressFile(dir, header.Name, r, os.FileMode(header.Mode))
			if e != nil {
				c.writeError(e)
				return
			}
		}
	}
	return
}
func (c *_UncompressWorker) uncompressDir(dir, name string, mode os.FileMode) (e error) {
	filename := filepath.Clean(dir + name)
	e = os.MkdirAll(filename, mode)
	return
}
func (c *_UncompressWorker) doneUncompressFile(dir, name string, r io.Reader, mode os.FileMode) (e error) {
	f, e := c.createFile(dir, name, mode)
	if e != nil {
		return
	}
	_, e = io.Copy(f, r)
	f.Close()
	return
}
func (c *_UncompressWorker) createFile(dir, name string, mode os.FileMode) (f *os.File, e error) {
	filename := filepath.Clean(dir + name)
	if c.style == CmdYesAll {
		f, e = os.Create(filename)
		return
	}
	if !os.IsExist(e) {
		return
	}
	if c.style == CmdSkipAll {
		e = nil
		return
	}

	e0 := c.writeExist(c.Name)
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
		f, e = os.Create(filename)
	} else if msg.Cmd == CmdYesAll {
		c.style = CmdYes
		f, e = os.Create(filename)
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
func (c *_UncompressWorker) uncompressZip(f *os.File, dir string) (e error) {
	fi, e := f.Stat()
	if e != nil {
		c.writeError(e)
		return
	}
	reader, e := zip.NewReader(f, fi.Size())
	if e != nil {
		c.writeError(e)
		return
	}

	for _, zipFile := range reader.File {
		name := zipFile.Name
		mode := zipFile.Mode()
		if mode.IsDir() {
			// 更新進度
			e = c.writeProgress(name)
			if e != nil {
				return
			}
			e = c.uncompressDir(dir, name, mode)
			if e != nil {
				c.writeError(e)
				break
			}
		} else {
			r, e0 := zipFile.Open()
			if e0 != nil {
				e = e0
				c.writeError(e)
				break
			}
			// 更新進度
			e = c.writeProgress(name)
			if e != nil {
				return
			}

			e = c.doneUncompressFile(dir, name, r, mode)
			r.Close()
			if e != nil {
				c.writeError(e)
				break
			}
		}
	}
	return
}
