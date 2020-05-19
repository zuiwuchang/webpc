package configure

import "path/filepath"

// Mount .
type Mount struct {
	// 網頁上 顯示的 目錄名稱
	Name string
	// 要映射的本地路徑
	Root string
	// 目錄是否可寫
	Write bool
}

// Format .
func (m *Mount) Format(basePath string) (e error) {
	if filepath.IsAbs(m.Root) {
		m.Root = filepath.Clean(m.Root)
	} else {
		m.Root = filepath.Clean(basePath + "/" + m.Root)
	}
	return
}
