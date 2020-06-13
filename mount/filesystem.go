package mount

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"gitlab.com/king011/webpc/logger"
	"go.uber.org/zap"
)

// Separator .
var Separator = string(filepath.Separator)

// FileSystem .
type FileSystem struct {
	ms []Mount
}

// Push .
func (f *FileSystem) Push(name, root string, read, write, shared bool) {
	f.ms = append(f.ms, Mount{
		name:   name,
		root:   root,
		read:   read,
		write:  write,
		shared: shared,
	})
}

// List .
func (f *FileSystem) List() []Mount {
	return f.ms
}

// Root .
func (f *FileSystem) Root(name string) *Mount {
	count := len(f.ms)
	for i := 0; i < count; i++ {
		if f.ms[i].name == name {
			return &f.ms[i]
		}
	}
	return nil
}

// Single .
func Single() *FileSystem {
	return &fs
}

// Mount .
type Mount struct {
	name                string
	root                string
	read, write, shared bool
}

// Name .
func (m *Mount) Name() string {
	return m.name
}

// Read .
func (m *Mount) Read() bool {
	return m.read
}

// Write .
func (m *Mount) Write() bool {
	return m.write
}

// Shared .
func (m *Mount) Shared() bool {
	return m.shared
}

// LS .
func (m *Mount) LS(path string) (dir string, results []FileInfo, e error) {
	dst, e := m.Filename(path)
	if e != nil {
		return
	}
	f, e := os.Open(dst)
	if e != nil {
		return
	}
	infos, e := f.Readdir(0)
	f.Close()
	count := len(infos)
	if e != nil {
		if count == 0 {
			return
		}
		if ce := logger.Logger.Check(zap.WarnLevel, "readdir error"); ce != nil {
			ce.Write(
				zap.Error(e),
			)
		}
		e = nil
	}

	dir = dst[len(m.root):]
	if Separator != `/` {
		dir = strings.ReplaceAll(path, Separator, `/`)
	}
	if !strings.HasPrefix(dir, `/`) {
		dir = `/` + dir
	}
	results = make([]FileInfo, count)
	for i := 0; i < count; i++ {
		results[i].Name = infos[i].Name()
		results[i].Mode = uint32(infos[i].Mode())
		results[i].Size = infos[i].Size()
		results[i].IsDir = infos[i].IsDir()
	}
	return
}

// Filename .
func (m *Mount) Filename(path string) (filename string, e error) {
	filename = filepath.Clean(m.root + path)
	if m.root != filename {
		root := m.root
		if !strings.HasSuffix(root, Separator) {
			root += Separator
		}
		if !strings.HasPrefix(filename, root) {
			e = errors.New(`Illegal path`)
			return
		}
	}
	return
}
