package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.com/king011/webpc/mount"
	"gitlab.com/king011/webpc/version"
	"gitlab.com/king011/webpc/web"
)

// Other 一些其它的 api
type Other struct {
	web.Helper
}

// Register impl IHelper
func (h Other) Register(router *gin.RouterGroup) {
	router.GET(`/version`, h.version)
	// 返回 用戶可訪問的 根目錄
	router.GET(`/roots`, h.roots)
}
func (h Other) version(c *gin.Context) {
	h.NegotiateData(c, http.StatusOK, gin.H{
		`tag`:    version.Tag,
		`commit`: version.Commit,
		`date`:   version.Date,
	})
}
func (h Other) roots(c *gin.Context) {
	session, _ := h.ShouldBindSession(c)

	var names []string
	fs := mount.Single()
	arrs := fs.List()
	for _, node := range arrs {
		if node.Shared() {
			names = append(names, node.Name())
			continue
		}
		if session == nil || !node.Read() {
			continue
		}
		if session.Root || session.Read || session.Write {
			names = append(names, node.Name())
		}
	}
	h.NegotiateData(c, http.StatusOK, names)
}
