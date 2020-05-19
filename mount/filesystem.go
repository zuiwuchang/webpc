package mount

// FileSystem .
type FileSystem struct {
	ms []Mount
}

// Push .
func (f *FileSystem) Push(name, root string, write bool) {
	f.ms = append(f.ms, Mount{
		name:  name,
		root:  root,
		write: write,
	})
}

// Single .
func Single() *FileSystem {
	return &fs
}

// Mount .
type Mount struct {
	name  string
	root  string
	write bool
}
