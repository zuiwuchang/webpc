package data

import (
	"bytes"
	"encoding/gob"
)

func init() {
	gob.Register(&User{})
}

// UserBucket .
const UserBucket = "user"

// User 用戶
type User struct {
	Name     string `json:"name,omitempty"`
	Password string `json:"-"`
	// 是否可獲取shell
	Shell bool `json:"shell,omitempty"`
	// 是否可讀取 檔案
	Read bool `json:"read,omitempty"`
	// 是否可寫入 檔案
	Write bool `json:"write,omitempty"`
	// 是否是 root
	Root bool `json:"root,omitempty"`
}

// Decode 由 []byte 解碼
func (u *User) Decode(b []byte) (e error) {
	decoder := gob.NewDecoder(bytes.NewBuffer(b))
	e = decoder.Decode(u)
	return
}

// Encoder 編碼到 []byte
func (u *User) Encoder() (b []byte, e error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	e = encoder.Encode(u)
	if e == nil {
		b = buffer.Bytes()
	}
	return
}
