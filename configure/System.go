package configure

import (
	"path/filepath"
	"strings"
)

// System 系統配置
type System struct {
	// 用戶數據庫
	DB string
	// 用戶shell
	Shell []string
	// 映射到web的目錄
	Mount []Mount
}

// Format .
func (s *System) Format(basePath string) (e error) {
	s.DB = strings.TrimSpace(s.DB)

	if filepath.IsAbs(s.DB) {
		s.DB = filepath.Clean(s.DB)
	} else {
		s.DB = filepath.Clean(basePath + "/" + s.DB)
	}

	for i := 0; i < len(s.Mount); i++ {
		e = s.Mount[i].Format(basePath)
		if e != nil {
			return
		}
	}
	return
}
