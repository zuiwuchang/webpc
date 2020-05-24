package mount

// FileInfo 檔案信息
type FileInfo struct {
	Name  string `json:"name,omitempty"`
	Dir   string `json:"dir,omitempty"`
	Mode  uint32 `json:"mode,omitempty"`
	Size  int64  `json:"size,omitempty"`
	IsDir bool   `json:"isDir,omitempty"`
}
