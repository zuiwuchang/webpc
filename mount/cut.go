package mount

import (
	"time"

	"github.com/gorilla/websocket"
)

// Cut 剪下檔案
func Cut(ws *websocket.Conn, dir, srcDir string, timeout time.Duration) (names []string, e error) {
	return
}
