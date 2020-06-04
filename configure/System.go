package configure

import (
	"path/filepath"
	"runtime"
	"strings"
)

// System 系統配置
type System struct {
	// 用戶數據庫
	DB string
	// 用戶shell 啓動腳本
	Shell string
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
	s.Shell = strings.TrimSpace(s.Shell)
	if s.Shell == "" {
		s.Shell = `shell-` + runtime.GOOS
		if runtime.GOOS == "windows" {
			s.Shell += `.bat`
		}
	}
	if filepath.IsAbs(s.Shell) {
		s.Shell = filepath.Clean(s.Shell)
	} else {
		s.Shell = filepath.Clean(basePath + "/" + s.Shell)
	}
	for i := 0; i < len(s.Mount); i++ {
		e = s.Mount[i].Format(basePath)
		if e != nil {
			return
		}
	}
	return
}
