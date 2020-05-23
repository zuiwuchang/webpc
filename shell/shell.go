package shell

import (
	"errors"
	"sync"

	"github.com/gorilla/websocket"
	"gitlab.com/king011/webpc/shell/internal/term"
)

const (
	// CmdError 錯誤
	CmdError = iota + 1
	// CmdResize 更改大小
	CmdResize
	// CmdInfo 返回终端信息
	CmdInfo
)

// ErrAlreadyAttach .
var ErrAlreadyAttach = errors.New(`shell already attach websocket`)

// Shell .
type Shell struct {
	term *term.Term
	conn *websocket.Conn

	username string
	shellid  int64
	name     string
	cols     uint16
	rows     uint16

	mutex sync.Mutex
}

// Run 运行 shell
func (s *Shell) Run(ws *websocket.Conn, cols, rows uint16) (e error) {
	// 運行 命令
	e = s.term.Start(cols, rows)
	if e != nil {
		return
	}
	s.cols = cols
	s.rows = rows

	s.conn = ws
	if ws != nil {
		ws.WriteMessage(websocket.BinaryMessage, []byte("\r\nwelcome guys, more info at https://gitlab.com/king011/webpc\r\n\r\n"))

		WriteInfo(ws, s.shellid, s.name)
	}

	// 等待進程結束
	go s.wait()
	// 讀取 tty
	go s.readTTY()
	return
}
func (s *Shell) wait() {
	s.term.Wait()

	s.term.Close()

	s.mutex.Lock()
	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}

	s.mutex.Unlock()
	// 更新數據庫 進程 結束
	Single().Unattach(s.username, s.shellid)
}

// IsAttack .
func (s *Shell) IsAttack() (yes bool) {
	s.mutex.Lock()
	yes = s.conn != nil
	s.mutex.Unlock()
	return
}

// Attack .
func (s *Shell) Attack(ws *websocket.Conn, cols, rows uint16) (e error) {
	s.mutex.Lock()
	if s.conn == nil {
		s.conn = ws
		if s.cols == cols && s.rows == rows {
			e0 := s.term.SetSize(cols+1, rows)
			if e0 == nil {
				e = s.term.SetSize(cols, rows)
			}
		} else {
			e = s.term.SetSize(cols, rows)
			s.cols = cols
			s.rows = rows
		}
		if e == nil {
			WriteInfo(ws, s.shellid, s.name)
		}
	} else {
		e = ErrAlreadyAttach
	}
	s.mutex.Unlock()
	return
}

// Unattack .
func (s *Shell) Unattack(ws *websocket.Conn) {
	s.mutex.Lock()
	if s.conn == ws {
		s.conn = nil
	}
	s.mutex.Unlock()
}
func (s *Shell) readTTY() {
	b := make([]byte, 1024)
	for {
		n, e := s.term.Read(b)
		if n != 0 {
			s.mutex.Lock()
			if s.conn != nil {
				e = s.conn.WriteMessage(websocket.BinaryMessage, b[:n])
				if e != nil {
					s.closeWebsocket()
				}
			}
			s.mutex.Unlock()
		}
		if e != nil {
			break
		}
	}
}
func (s *Shell) closeWebsocket() {
	s.conn.Close()
	s.conn = nil
}

// Kill 關閉 進程
func (s *Shell) Kill() {
	s.mutex.Lock()
	s.term.Kill()
	s.mutex.Unlock()
}

// SetSize .
func (s *Shell) SetSize(cols, rows uint16) (e error) {
	s.mutex.Lock()
	e = s.term.SetSize(cols, rows)
	s.mutex.Unlock()
	return
}

// Write .
func (s *Shell) Write(b []byte) (int, error) {
	return s.term.Write(b)
}
