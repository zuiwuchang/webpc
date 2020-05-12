package utils

import (
	"bytes"
	"fmt"
)

// Command client command
type Command struct {
	Name  string
	Usage string
	Done  func() bool
}

// Management command management
type Management struct {
	keys map[string]Command
	cmd  []Command
}

// AddCommand add command to management
func (m *Management) AddCommand(cmds ...Command) {
	if m.keys == nil {
		m.keys = make(map[string]Command)
	}
	if m.cmd == nil {
		m.cmd = make([]Command, 0, len(cmds))
	}
	for i := 0; i < len(cmds); i++ {
		cmd := cmds[i]
		name := cmd.Name
		if _, ok := m.keys[name]; ok {
			panic(fmt.Sprintf("command already exists : %v", name))
		}
		m.keys[name] = cmd
		m.cmd = append(m.cmd, cmd)
	}
}

// Done .
func (m *Management) Done(str string) (exit bool, e error) {
	if cmd, ok := m.keys[str]; ok {
		exit = cmd.Done()
		return
	}
	e = fmt.Errorf("not support command : %v", str)
	return
}

// Usage .
func (m *Management) Usage() string {
	var buf bytes.Buffer

	max := 0
	for i := 0; i < len(m.cmd); i++ {
		val := len(m.cmd[i].Name)
		if max < val {
			max = val
		}
	}
	for i := 0; i < len(m.cmd); i++ {
		format := fmt.Sprintf("%%-%vv %%v\n", max+4)
		buf.WriteString(fmt.Sprintf(format, m.cmd[i].Name, m.cmd[i].Usage))
	}

	return buf.String()
}
