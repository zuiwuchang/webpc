package mount

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gitlab.com/king011/webpc/helper"
)

// CompressMessage .
type CompressMessage struct {
	Cmd   int      `json:"cmd,omitempty"`
	Name  string   `json:"name,omitempty"`
	Names []string `json:"names,omitempty"`
}

// Compress 壓縮
func (m *Mount) Compress(ws *websocket.Conn, dir string, timeout time.Duration) (name string, names []string, e error) {
	timer := time.NewTimer(timeout)
	ch := make(chan CompressMessage)
	var errStr string
	go func() {
		for {
			t, b, e := ws.ReadMessage()
			if e != nil {
				errStr = e.Error()
				break
			} else if t != websocket.TextMessage {
				errStr = `not support message type`
				break
			}
			var msg CompressMessage
			e = helper.Unmarshal(b, &msg)
			if e != nil {
				errStr = e.Error()
				break
			}
			if msg.Cmd == CmdHeart {
				continue
			} else if msg.Cmd == CmdInit {
				ch <- msg
				return
			} else {
				errStr = fmt.Sprint(`not support cmd `, msg.Cmd)
				break
			}
		}
		close(ch)
	}()
	var msg CompressMessage
	select {
	case <-timer.C:
		helper.WSWriteJSON(ws, gin.H{
			`cmd`:   CmdError,
			`error`: `wait init timeout`,
		})
		ws.Close()
		<-ch
		e = context.DeadlineExceeded
		return
	case msg = <-ch:
		if msg.Cmd != CmdInit {
			helper.WSWriteJSON(ws, gin.H{
				`cmd`:   CmdError,
				`error`: errStr,
			})
			e = errors.New(errStr)
			return
		}
	}

	e = m.compress(ws, dir, msg.Name, msg.Names)
	if e != nil {
		return
	}
	name = msg.Name
	names = msg.Names
	return
}

func (m *Mount) compress(ws *websocket.Conn, dir, name string, names []string) (e error) {
	count := len(names)
	if count == 0 {
		e = fmt.Errorf(`names nil`)
		helper.WSWriteJSON(ws, gin.H{
			`cmd`:   CmdError,
			`error`: e.Error(),
		})
		return
	}

	if !strings.HasSuffix(dir, Separator) {
		dir += Separator
	}
	dst := filepath.Clean(dir + name)
	if !strings.HasPrefix(dst, dir) {
		e = fmt.Errorf(`name not support : %v`, name)
		helper.WSWriteJSON(ws, gin.H{
			`cmd`:   CmdError,
			`error`: e.Error(),
		})
		return
	}
	str := filepath.Base(dst)
	if str != name || str == `.` || str == `..` {
		e = fmt.Errorf(`name not support : %v`, name)
		helper.WSWriteJSON(ws, gin.H{
			`cmd`:   CmdError,
			`error`: e.Error(),
		})
		return
	}

	f, e := os.Create(dst)
	if e != nil {
		helper.WSWriteJSON(ws, gin.H{
			`cmd`:   CmdError,
			`error`: e.Error(),
		})
		return
	}
	defer func() {
		if f != nil {
			f.Close()
			os.Remove(dst)
		}
	}()

	for i := 0; i < count; i++ {
		str := filepath.Clean(dir + names[i])
		if !strings.HasPrefix(str, dir) {
			e = fmt.Errorf(`name not support : %v`, names[i])
			helper.WSWriteJSON(ws, gin.H{
				`cmd`:   CmdError,
				`error`: e.Error(),
			})
			return
		}
		str = filepath.Base(str)
		if str != names[i] || str == `.` || str == `..` || name == str {
			e = fmt.Errorf(`name not support : %v`, names[i])
			helper.WSWriteJSON(ws, gin.H{
				`cmd`:   CmdError,
				`error`: e.Error(),
			})
			return
		}
	}
	e = m.archive(ws, f, dir, names)
	if e != nil {
		return
	}
	return
}
func (m *Mount) archive(ws *websocket.Conn, f *os.File, dir string, names []string) (e error) {
	gf := gzip.NewWriter(f)
	defer gf.Close()

	return
}
