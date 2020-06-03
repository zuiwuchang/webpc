package data

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
)

func init() {
	gob.Register(&Shell{})
}

// ShellBucket .
const ShellBucket = "shell"

// Shell 启动的 shell 用于 重启服务时 恢复shell
type Shell struct {
	ID         int64  `json:"id,omitempty"`
	Name       string `json:"name,omitempty"`
	FontSize   int    `json:"fontSize,omitempty"`
	FontFamily string `json:"fontFamily,omitempty"`
}

// EncoderID .
func EncoderID(id uint64) (b []byte) {
	b = make([]byte, 8)
	binary.LittleEndian.PutUint64(b, id)
	return
}

// DecodeID .
func DecodeID(b []byte) (id uint64) {
	if len(b) >= 8 {
		id = binary.LittleEndian.Uint64(b)
	}
	return
}

// Decode 由 []byte 解碼
func (s *Shell) Decode(b []byte) (e error) {
	decoder := gob.NewDecoder(bytes.NewBuffer(b))
	e = decoder.Decode(s)
	return
}

// Encoder 編碼到 []byte
func (s *Shell) Encoder() (b []byte, e error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	e = encoder.Encode(s)
	if e == nil {
		b = buffer.Bytes()
	}
	return
}
