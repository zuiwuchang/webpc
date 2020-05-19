package cookie

import (
	"encoding/gob"
)

const (
	// CookieName cookie key name
	CookieName = "golang-webpc-session"
)

func init() {
	gob.Register(&Session{})
}

// Session user session info
type Session struct {
	Name string `json:"name,omitempty"`
	// 是否可獲取shell
	Shell bool `json:"shell,omitempty"`
	// 是否可讀取 檔案
	Read bool `json:"read,omitempty"`
	// 是否可寫入 檔案
	Write bool `json:"write,omitempty"`
	// 是否是 root
	Root bool `json:"root,omitempty"`
}

// Cookie encode to cookie
func (s *Session) Cookie() (string, error) {
	return Encode("session", s)
}

// IsRoot if user is root return true
func (s *Session) IsRoot() (yes bool) {
	yes = s.Root
	return
}

// FromCookie restore session from cookie
func FromCookie(val string) (session *Session, e error) {
	var s Session
	e = Decode("session", val, &s)
	if e != nil {
		return
	}
	session = &s
	return
}
