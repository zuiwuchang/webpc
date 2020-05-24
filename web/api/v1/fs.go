package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.com/king011/webpc/mount"
	"gitlab.com/king011/webpc/web"
)

// FS 檔案系統
type FS struct {
	web.Helper
}

// Register impl IHelper
func (h FS) Register(router *gin.RouterGroup) {
	r := router.Group(`fs`)

	r.GET(`ls`, h.ls)
	r.POST(`ls`, h.ls)
}
func (h FS) ls(c *gin.Context) {
	var obj struct {
		Root string `form:"root" json:"root" xml:"root" yaml:"root" binding:"required"`
		Path string `form:"path" json:"path" xml:"path" yaml:"path" binding:"required"`
	}
	e := h.Bind(c, &obj)
	if e != nil {
		return
	}
	fs := mount.Single()
	m := fs.Root(obj.Root)
	if m == nil {
		c.Status(http.StatusNotFound)
		return
	}
	if !h.checkRead(c, m) {
		return
	}
}
func (h FS) checkRead(c *gin.Context, m *mount.Mount) (ok bool) {
	if m.Shared() {
		ok = true
		return
	}
	session := h.BindSession(c)
	if session.Root {
		ok = true
		return
	}
	if !m.Read() {
		c.Status(http.StatusForbidden)
		return
	}

	ok = true
	return
}
