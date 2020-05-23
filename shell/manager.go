package shell

import (
	"errors"
	"os"
	"sync"
	"time"

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

// ListInfo .
type ListInfo struct {
	// shell id
	ID int64 `json:"id,omitempty"`
	// shell 顯示名稱
	Name string `json:"name,omitempty"`
	// 是否 附加 websocket
	Attached bool `json:"attached,omitempty"`
}

// Manager .
type Manager struct {
	mutex sync.Mutex
	keys  map[string]*Element
}

// Unattach 進程結束 釋放資源
func (m *Manager) Unattach(username string, shellid int64) (e error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	element, ok := m.keys[username]
	if !ok {
		return
	}
	element.Unattach(username, shellid)
	return
}

// Attach .
func (m *Manager) Attach(ws *websocket.Conn, username string, shellid int64, cols, rows uint16, newshell bool) (s *Shell, e error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	element, ok := m.keys[username]
	if !ok {
		element = &Element{
			keys: make(map[int64]*Shell),
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

// List 列举 shell 状态
func (m *Manager) List(username string) (arrs []ListInfo) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if element, ok := m.keys[username]; ok {
		for _, s := range element.keys {
			arrs = append(arrs, ListInfo{
				ID:       s.shellid,
				Name:     s.name,
				Attached: s.IsAttack(),
			})
		}
	}
	return
}

// Element .
type Element struct {
	keys map[int64]*Shell
}

// Attach .
func (element *Element) Attach(ws *websocket.Conn, username string, shellid int64, cols, rows uint16, newshell bool) (s *Shell, e error) {
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
			name:     time.Unix(shellid, 0).Local().Format(`2006/01/02 15:04:05`),
		}
		e = shell.Run(ws, username, shellid, cols, rows)
		if e != nil {
			return
		}
		s = shell
		element.keys[shellid] = s
		// 更新數據庫
		element.add(username, shellid, shell.name)
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

// Unattach .
func (element *Element) Unattach(username string, shellid int64) (ok bool) {
	_, ok = element.keys[shellid]
	if !ok {
		return
	}
	delete(element.keys, shellid)
	element.remove(username, shellid)
	return
}
func (element *Element) add(username string, shellid int64, name string) {
	// 更新數據庫
}
func (element *Element) remove(username string, shellid int64) {
	// 更新數據庫
}
