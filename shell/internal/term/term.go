package term

import (
	"os"
	"os/exec"

	"github.com/creack/pty"
)

// Term .
type Term struct {
	name string
	args []string
	cmd  *exec.Cmd
	tty  *os.File
}

// New .
func New(name string, args ...string) *Term {
	return &Term{
		name: name,
		args: args,
	}
}

// Start 運行 命令
func (t *Term) Start(cols, rows uint16) (e error) {
	cmd := exec.Command(t.name, t.args...)
	f, e := pty.StartWithSize(cmd, &pty.Winsize{
		Cols: cols,
		Rows: rows,
	})
	if e != nil {
		return
	}
	t.cmd = cmd
	t.tty = f
	return
}

// Kill 關閉進程
func (t *Term) Kill() (e error) {
	return t.cmd.Process.Kill()
}

// Wait .
func (t *Term) Wait() error {
	return t.cmd.Wait()
}

// Close .
func (t *Term) Close() error {
	return t.tty.Close()
}

// Read .
func (t *Term) Read(b []byte) (int, error) {
	return t.tty.Read(b)
}

// Write .
func (t *Term) Write(b []byte) (int, error) {
	return t.tty.Write(b)
}

// SetSize .
func (t *Term) SetSize(cols, rows uint16) error {
	return pty.Setsize(t.tty, &pty.Winsize{
		Cols: cols,
		Rows: rows,
	})
}
