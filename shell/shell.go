package shell

import (
	"errors"
	"sync"

	"github.com/gorilla/websocket"
	"gitlab.com/king011/webpc/shell/internal/term"
)

const (
	// DataTypeTTY tty 消息
	DataTypeTTY = iota + 1
	// DataTypeError 錯誤
	DataTypeError
	// DataTypeResize 更改大小
	DataTypeResize
)

// ErrAlreadyAttach .
var ErrAlreadyAttach = errors.New(`shell already attach websocket`)

// Shell .
type Shell struct {
	term *term.Term
	conn *websocket.Conn

	username string
	shellid  string
	name     string

	mutex sync.Mutex
}

// New 创建 shell
func New(name string, args ...string) (shell *Shell, e error) {
	shell = &Shell{
		term: term.New(name, args...),
	}
	return
}

// Run 运行 shell
func (s *Shell) Run(ws *websocket.Conn, username, shellid string, cols, rows uint16) (e error) {
	// 運行 命令
	e = s.term.Start(cols, rows)
	if e != nil {
		return
	}

	s.conn = ws
	if ws != nil {
		ws.WriteMessage(websocket.BinaryMessage, []byte("\r\nwelcome guys, more info at https://gitlab.com/king011/webpc\r\n\r\n"))
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

// Attack .
func (s *Shell) Attack(ws *websocket.Conn, cols, rows uint16) (e error) {
	s.mutex.Lock()
	if s.conn == nil {
		s.conn = ws
		e = s.term.SetSize(cols, rows)
	} else {
		e = ErrAlreadyAttach
	}
	s.mutex.Unlock()
	return
}
func (s *Shell) readTTY() {
	b := make([]byte, 1024)
	for {
		n, e := s.term.Read(b[1:])
		if n != 0 {
			s.mutex.Lock()
			if s.conn != nil {
				b[0] = DataTypeTTY
				e = s.conn.WriteMessage(websocket.BinaryMessage, b[:1+n])
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
