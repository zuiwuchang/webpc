package shell

import (
	"os"
	"os/exec"

	"github.com/gorilla/websocket"
)

// Shell .
type Shell struct {
	cmd  *exec.Cmd
	tty  *os.File
	conn *websocket.Conn
}

// New 创建 shell
func New(name string, args ...string) (shell *Shell, e error) {
	cmd := exec.Command(name, args...)
	shell = &Shell{
		cmd: cmd,
	}
	return
}

// Run 运行 shell
func (s *Shell) Run() {

}
