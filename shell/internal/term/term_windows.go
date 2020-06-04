// +build windows

package term

import (
	"os"
	"sync"
	"syscall"

	"github.com/iamacarpet/go-winpty"
	"gitlab.com/king011/webpc/utils"
)

// Term .
type Term struct {
	name   string
	args   []string
	tty    *winpty.WinPTY
	handle uintptr
	mutex  sync.Mutex
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
	env := os.Environ()
	env = append(env, `ShellUser=`+t.args[0])
	tty, e := winpty.OpenWithOptions(winpty.Options{
		DLLPrefix: utils.BasePath(),
		Command:   t.name,
		Env:       env,
	})
	if e != nil {
		return
	}
	t.tty = tty
	t.handle = tty.GetProcHandle()
	t.SetSize(cols, rows)
	return
}

// Kill 關閉進程
func (t *Term) Kill() (e error) {
	t.mutex.Lock()
	t.tty.Close()
	t.mutex.Unlock()
	return
}

// Wait .
func (t *Term) Wait() error {
	syscall.WaitForSingleObject(syscall.Handle(t.handle), syscall.INFINITE)
	return nil
}

// Close .
func (t *Term) Close() error {
	t.mutex.Lock()
	t.tty.Close()
	t.mutex.Unlock()
	return nil
}

// Read .
func (t *Term) Read(b []byte) (int, error) {
	n, e := t.tty.StdOut.Read(b)
	if e != nil {
		t.Close()
	}
	return n, e
}

// Write .
func (t *Term) Write(b []byte) (int, error) {
	n, e := t.tty.StdIn.Write(b)
	if e != nil {
		t.Close()
	}
	return n, e
}

// SetSize .
func (t *Term) SetSize(cols, rows uint16) error {
	t.tty.SetSize(uint32(cols), uint32(rows))
	return nil
}
