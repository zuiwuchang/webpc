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
