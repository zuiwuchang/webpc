package shell

import (
	"time"

	"github.com/gorilla/websocket"
	"gitlab.com/king011/webpc/db/data"
	"gitlab.com/king011/webpc/db/manipulator"
)

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
		shell := &Shell{
			term:     newTerm(),
			username: username,
			shellid:  shellid,
			name:     time.Unix(shellid, 0).Local().Format(`2006/01/02 15:04:05`),
		}
		e = shell.Run(ws, cols, rows)
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
	var mShell manipulator.Shell
	mShell.Add(username, &data.Shell{
		ID:   shellid,
		Name: name,
	})
}
func (element *Element) remove(username string, shellid int64) {
	// 更新數據庫
	var mShell manipulator.Shell
	mShell.Remove(username, shellid)
}

// Rename .
func (element *Element) Rename(username string, shellid int64, name string) (e error) {
	s, ok := element.keys[shellid]
	if !ok {
		e = ErrShellidNotExists
		return
	}
	s.Rename(name)
	// 更新 数据库
	var mShell manipulator.Shell
	mShell.Rename(username, shellid, name)
	return
}

// Kill .
func (element *Element) Kill(username string, shellid int64) (e error) {
	s, ok := element.keys[shellid]
	if !ok {
		e = ErrShellidNotExists
		return
	}
	s.Kill()
	return
}
