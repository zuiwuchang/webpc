package configure

import "path/filepath"

// Mount .
type Mount struct {
	// 網頁上 顯示的 目錄名稱
	Name string
	// 要映射的本地路徑
	Root string

	// 設置目錄可讀 有讀取/寫入權限的用戶 可以 讀取檔案
	Read bool

	// 設置目錄可寫 有寫入權限的用戶可以 寫入檔案
	// 如果 Write 爲 true 則 Read 會被強制設置爲 true
	Write bool

	// 設置爲共享目錄 允許任何人讀取檔案
	// 如果 Shared 爲 true 則 Read 會被強制設置爲 true
	Shared bool
}

// Format .
func (m *Mount) Format(basePath string) (e error) {
	if filepath.IsAbs(m.Root) {
		m.Root = filepath.Clean(m.Root)
	} else {
		m.Root = filepath.Clean(basePath + "/" + m.Root)
	}

	if m.Write || m.Shared {
		m.Read = true
	}
	return
}
