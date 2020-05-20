package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
}
func (h Other) version(c *gin.Context) {
	h.NegotiateData(c, http.StatusOK, gin.H{
		`tag`:    version.Tag,
		`commit`: version.Commit,
		`date`:   version.Date,
	})
}
