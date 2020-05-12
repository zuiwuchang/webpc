package configure

import (
	"path/filepath"
	"strings"
)

// HTTP configure http
type HTTP struct {
	Addr string

	CertFile string
	KeyFile  string
}

// TLS if tls return true
func (c *HTTP) TLS() bool {
	return c.CertFile != "" && c.KeyFile != ""
}

// Format .
func (c *HTTP) Format(basePath string) (e error) {
	c.Addr = strings.TrimSpace(c.Addr)
	c.CertFile = strings.TrimSpace(c.CertFile)
	c.KeyFile = strings.TrimSpace(c.KeyFile)

	if c.TLS() {
		if filepath.IsAbs(c.CertFile) {
			c.CertFile = filepath.Clean(c.CertFile)
		} else {
			c.CertFile = filepath.Clean(basePath + "/" + c.CertFile)
		}

		if filepath.IsAbs(c.KeyFile) {
			c.KeyFile = filepath.Clean(c.KeyFile)
		} else {
			c.KeyFile = filepath.Clean(basePath + "/" + c.KeyFile)
		}
	}
	return
}
