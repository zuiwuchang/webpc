package shell

import (
	"errors"
	"runtime"
	"sync"

	"github.com/boltdb/bolt"
	"github.com/gorilla/websocket"
	"gitlab.com/king011/webpc/configure"
	"gitlab.com/king011/webpc/db/data"
	"gitlab.com/king011/webpc/db/manipulator"
	"gitlab.com/king011/webpc/logger"
	"gitlab.com/king011/webpc/shell/internal/term"
	"gitlab.com/king011/webpc/utils"
	"go.uber.org/zap"
)

// ErrShellidDuplicate .
var ErrShellidDuplicate = errors.New(`new shell id duplicate`)

// ErrShellidNotExists .
var ErrShellidNotExists = errors.New(`shell id not exists`)

// ErrUsernameNotExists .
var ErrUsernameNotExists = errors.New(`shell username not exists`)

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

// Restore 恢复 shell
func (m *Manager) Restore() {
	manipulator.DB().View(func(t *bolt.Tx) (e error) {
		bucket := t.Bucket(utils.StringToBytes(data.ShellBucket))
		if bucket == nil {
			return
		}
		cursor := bucket.Cursor()
		for k, _ := cursor.First(); k != nil; k, _ = cursor.Next() {
			shell := bucket.Bucket(k)
			if shell == nil {
				return
			}
			m.restore(string(k), shell)
		}
		return
	})
	return
}
func newTerm(username string) *term.Term {
	cnf := configure.Single()
	return term.New(cnf.System.Shell, username, runtime.GOOS, runtime.GOARCH)
}
func (m *Manager) restore(username string, bucket *bolt.Bucket) {
	cursor := bucket.Cursor()
	for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
		var shell data.Shell
		e := shell.Decode(v)
		if e != nil {
			if ce := logger.Logger.Check(zap.ErrorLevel, `restore shell error`); ce != nil {
				ce.Write(
					zap.Error(e),
					zap.String(`username`, username),
					zap.Int64(`key`, int64(data.DecodeID(k))),
				)
			}
			continue
		}
		s := &Shell{
			term:       newTerm(username),
			username:   username,
			shellid:    shell.ID,
			name:       shell.Name,
			fontSize:   shell.FontSize,
			fontFamily: shell.FontFamily,
		}
		e = s.Run(nil, 24, 10)
		if e != nil {
			if ce := logger.Logger.Check(zap.ErrorLevel, `restore shell error`); ce != nil {
				ce.Write(
					zap.Error(e),
					zap.String(`username`, username),
					zap.Int64(`key`, shell.ID),
					zap.String(`name`, shell.Name),
				)
			}
			continue
		}

		if ce := logger.Logger.Check(zap.InfoLevel, `restore shell`); ce != nil {
			ce.Write(
				zap.String(`username`, username),
				zap.Int64(`key`, shell.ID),
				zap.String(`name`, shell.Name),
			)
		}

		element := m.keys[username]
		if element == nil {
			element = &Element{
				keys: make(map[int64]*Shell),
			}
			m.keys[username] = element
		}
		element.keys[s.shellid] = s
	}
}

// Rename .
func (m *Manager) Rename(username string, shellid int64, name string) (e error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	element, ok := m.keys[username]
	if !ok {
		e = ErrUsernameNotExists
		return
	}

	e = element.Rename(username, shellid, name)
	return
}

// Kill .
func (m *Manager) Kill(username string, shellid int64) (e error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	element, ok := m.keys[username]
	if !ok {
		e = ErrUsernameNotExists
		return
	}

	e = element.Kill(username, shellid)
	return
}
