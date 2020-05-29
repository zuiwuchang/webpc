package mount

import (
	"archive/tar"
	"archive/zip"
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

const algorithmTarGZ = 0
const algorithmTar = 1
const algorithmZip = 2

// Compress 執行 壓縮
func Compress(ws *websocket.Conn, dir string, timeout time.Duration) (name string, names []string, e error) {
	c := &_CompressWorker{
		ws:     ws,
		ch:     make(chan _CompressMessage),
		cancel: make(chan struct{}),
	}
	e = c.done(dir, timeout)
	c.close()
	name = c.Name
	names = c.Names
	return
}

type _CompressMessage struct {
	Cmd       int      `json:"cmd,omitempty"`
	Name      string   `json:"name,omitempty"`
	Names     []string `json:"names,omitempty"`
	Algorithm int      `json:"algorithm,omitempty"`
}

type _CompressWorker struct {
	Name      string
	Names     []string
	ws        *websocket.Conn
	ch        chan _CompressMessage
	cancel    chan struct{}
	closed    int32
	Algorithm int
}

func (c *_CompressWorker) close() {
	if atomic.CompareAndSwapInt32(&c.closed, 0, 1) {
		c.ws.Close()
		close(c.cancel)
	}
}
func (c *_CompressWorker) read() {
	for {
		t, b, e := c.ws.ReadMessage()
		if e != nil {
			break
		} else if t != websocket.TextMessage {
			if ce := logger.Logger.Check(zap.WarnLevel, `compress not support message type`); ce != nil {
				ce.Write(
					zap.Int("type", t),
				)
			}
			break
		}
		var msg _CompressMessage
		e = helper.Unmarshal(b, &msg)
		if e != nil {
			if ce := logger.Logger.Check(zap.WarnLevel, `compress unmarshal message error`); ce != nil {
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
func (c *_CompressWorker) getMessage(timeout time.Duration) (msg _CompressMessage, e error) {
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
func (c *_CompressWorker) waitMessage() (msg _CompressMessage, e error) {
	select {
	case <-c.cancel:
		e = context.Canceled
	case msg = <-c.ch:
	}
	return
}
func (c *_CompressWorker) done(dir string, timeout time.Duration) (e error) {
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
	c.Names = msg.Names
	c.Algorithm = msg.Algorithm
	switch c.Algorithm {
	case algorithmTar:
		if !strings.HasSuffix(strings.ToLower(filepath.Base(c.Name)), `.tar`) {
			c.Name += `.tar`
		}
	case algorithmZip:
		if !strings.HasSuffix(strings.ToLower(filepath.Base(c.Name)), `.zip`) {
			c.Name += `.zip`
		}
	default:
		if !strings.HasSuffix(strings.ToLower(filepath.Base(c.Name)), `.tar.gz`) {
			c.Name += `.tar.gz`
		}
	}
	info, e := c.compress(dir)
	if e != nil {
		return
	}
	c.writeDone(info)
	return
}
func (c *_CompressWorker) writeExist(val string) error {
	return helper.WSWriteJSON(c.ws, gin.H{
		`cmd`: CmdExist,
		`val`: val,
	})
}
func (c *_CompressWorker) writeError(e error) error {
	return helper.WSWriteJSON(c.ws, gin.H{
		`cmd`:   CmdError,
		`error`: e.Error(),
	})
}
func (c *_CompressWorker) writeDone(info *FileInfo) error {
	return helper.WSWriteJSON(c.ws, gin.H{
		`cmd`:      CmdDone,
		`fileInfo`: info,
	})
}
func (c *_CompressWorker) writeProgress(val string) error {
	return helper.WSWriteJSON(c.ws, gin.H{
		`cmd`: CmdProgress,
		`val`: val,
	})
}
func (c *_CompressWorker) compress(dir string) (result *FileInfo, e error) {
	count := len(c.Names)
	if count == 0 {
		e = errors.New(`names nil`)
		c.writeError(e)
		return
	}

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
	f, e := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	if e != nil {
		if os.IsExist(e) {
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
				f, e = os.Create(dst)
			} else {
				e = fmt.Errorf(`Expect yes or no,Unexpected instruction %v`, msg.Cmd)
				c.writeError(e)
				return
			}
		}
		if e != nil {
			c.writeError(e)
			return
		}
	}
	defer func() {
		if f != nil {
			f.Close()
			os.Remove(dst)
		}
	}()

	for i := 0; i < count; i++ {
		str := filepath.Clean(dir + c.Names[i])
		if !strings.HasPrefix(str, dir) {
			e = fmt.Errorf(`name not support : %v`, c.Names[i])
			c.writeError(e)
			return
		}
		str = filepath.Base(str)
		if str != c.Names[i] || str == `.` || str == `..` || c.Name == str {
			e = fmt.Errorf(`name not support : %v`, c.Names[i])
			c.writeError(e)
			return
		}
	}
	switch c.Algorithm {
	case algorithmZip:
		e = c.archiveZip(f, dir)
	case algorithmTar:
		e = c.archiveTar(f, dir)
	default:
		e = c.archiveTarGz(f, dir)
	}
	if e != nil {
		return
	}
	ret, _ := f.Seek(0, os.SEEK_END)
	f.Close()
	f = nil
	result = &FileInfo{
		Name:  c.Name,
		Size:  ret,
		Mode:  uint32(0666),
		IsDir: false,
	}
	return
}
func (c *_CompressWorker) archiveTarGz(f *os.File, dir string) (e error) {
	gz := gzip.NewWriter(f)
	w := tar.NewWriter(gz)
	for _, name := range c.Names {
		e = c.archiveTarRoot(w, dir, name)
		if e != nil {
			break
		}
	}
	w.Close()
	gz.Close()
	return
}
func (c *_CompressWorker) archiveTar(f *os.File, dir string) (e error) {
	w := tar.NewWriter(f)
	for _, name := range c.Names {
		e = c.archiveTarRoot(w, dir, name)
		if e != nil {
			break
		}
	}
	w.Close()
	return
}
func (c *_CompressWorker) archiveTarRoot(w *tar.Writer, dir string, name string) (e error) {
	root := filepath.Clean(dir + name)
	count := len(dir)
	e = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			c.writeError(err)
			return err
		}
		name := path[count:]
		e := c.writeProgress(name)
		if e != nil {
			return e
		}
		if info.IsDir() {
			e = c.archiveTarDir(w, info, name)
		} else {
			e = c.archiveTarFile(w, info, path, name)
		}
		if e != nil {
			c.writeError(e)
		}
		return e
	})
	return
}

func (c *_CompressWorker) archiveTarDir(w *tar.Writer, info os.FileInfo, name string) error {
	header := &tar.Header{
		Typeflag: tar.TypeDir,
		Name:     name,
		Mode:     int64(info.Mode()),
		Uid:      os.Getuid(),
		Gid:      os.Getgid(),
		Size:     info.Size(),
		ModTime:  info.ModTime(),
	}
	return w.WriteHeader(header)
}
func (c *_CompressWorker) archiveTarFile(w *tar.Writer, info os.FileInfo, filename, name string) (e error) {
	f, e := os.Open(filename)
	if e != nil {
		return
	}
	defer f.Close()
	info, e = f.Stat()
	if e != nil {
		return
	}
	if info.IsDir() {
		e = c.archiveTarDir(w, info, name)
		return
	}

	header := &tar.Header{
		Name:    name,
		Mode:    int64(info.Mode()),
		Uid:     os.Getuid(),
		Gid:     os.Getgid(),
		Size:    info.Size(),
		ModTime: info.ModTime(),
	}
	if e = w.WriteHeader(header); e != nil {
		return
	}
	_, e = io.Copy(w, f)
	return
}
func (c *_CompressWorker) archiveZip(f *os.File, dir string) (e error) {
	w := zip.NewWriter(f)
	for _, name := range c.Names {
		e = c.archiveZipRoot(w, dir, name)
		if e != nil {
			break
		}
	}
	w.Close()
	return
}
func (c *_CompressWorker) archiveZipRoot(w *zip.Writer, dir string, name string) (e error) {
	root := filepath.Clean(dir + name)
	count := len(dir)
	e = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			c.writeError(err)
			return err
		}
		name := path[count:]
		e := c.writeProgress(name)
		if e != nil {
			return e
		}
		if info.IsDir() {
			e = c.archiveZipDir(w, info, name)
		} else {
			e = c.archiveZipFile(w, info, path, name)
		}
		if e != nil {
			c.writeError(e)
		}
		return e
	})
	return
}
func (c *_CompressWorker) archiveZipDir(w *zip.Writer, info os.FileInfo, name string) error {
	header, e := zip.FileInfoHeader(info)
	if e != nil {
		return e
	}
	header.Name = name
	_, e = w.CreateHeader(header)
	return e
}
func (c *_CompressWorker) archiveZipFile(w *zip.Writer, info os.FileInfo, filename, name string) (e error) {
	f, e := os.Open(filename)
	if e != nil {
		return
	}
	defer f.Close()
	info, e = f.Stat()
	if e != nil {
		return
	}
	if info.IsDir() {
		e = c.archiveZipDir(w, info, name)
		return
	}

	header, e := zip.FileInfoHeader(info)
	if e != nil {
		return
	}
	header.Name = name
	zw, e := w.CreateHeader(header)
	if e != nil {
		return
	}
	_, e = io.Copy(zw, f)
	return
}
