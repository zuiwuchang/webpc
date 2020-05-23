package shell

import (
	"errors"
	"os"
	"sync"

	"github.com/gorilla/websocket"
	"gitlab.com/king011/webpc/configure"
	"gitlab.com/king011/webpc/shell/internal/term"
)

// ErrShellidDuplicate .
var ErrShellidDuplicate = errors.New(`new shell id duplicate`)

// ErrShellidNotExists .
var ErrShellidNotExists = errors.New(`attach shell id not exists`)

var manager = Manager{
	keys: make(map[string]*Element),
}

// Single .
func Single() *Manager {
	return &manager
}

// Manager .
type Manager struct {
	mutex sync.Mutex
	keys  map[string]*Element
}

func (m *Manager) unattach(username, shellid string) (e error) {
	return
}

// Attach .
func (m *Manager) Attach(ws *websocket.Conn, username, shellid string, cols, rows uint16, newshell bool) (s *Shell, e error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	element, ok := m.keys[username]
	if !ok {
		element = &Element{
			keys: make(map[string]*Shell),
		}
	}

	s, e = element.Attach(ws, username, shellid, cols, rows, newshell)
	if e != nil {
		return
	}

	if !ok {
		m.keys[username] = element
	}
	return
}

// Element .
type Element struct {
	keys map[string]*Shell
}

// Attach .
func (element *Element) Attach(ws *websocket.Conn, username, shellid string, cols, rows uint16, newshell bool) (s *Shell, e error) {
	if newshell {
		if _, ok := element.keys[shellid]; ok {
			e = ErrShellidDuplicate
			return
		}
		cnf := configure.Single()
		var name string
		var args []string
		count := len(cnf.System.Shell)
		if count == 0 {
			name = os.Getenv(`SHELL`)
		} else {
			name = cnf.System.Shell[0]
		}
		if count > 1 {
			args = cnf.System.Shell[1:]
		}
		shell := &Shell{
			term:     term.New(name, args...),
			username: username,
			shellid:  shellid,
			name:     shellid,
		}
		e = shell.Run(ws, username, shellid, cols, rows)
		if e != nil {
			return
		}
		s = shell
		// 更新數據庫
		element.add(username, shellid, shellid)
	} else {
		shell, ok := element.keys[shellid]
		if !ok {
			e = ErrShellidNotExists
			return
		}
		e = shell.Attack(ws, cols, rows)
		if e != nil {
			return
		}
		s = shell
	}
	return
}

func (element *Element) add(username, shellid, name string) {

}
