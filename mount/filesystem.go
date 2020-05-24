package mount

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
