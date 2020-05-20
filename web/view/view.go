package view

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/king011/webpc/web"
)

// BaseURL request base url
const BaseURL = `/view`

// Helper path of /app
type Helper struct {
	web.Helper
}

// Register impl IHelper
func (h Helper) Register(router *gin.RouterGroup) {
	router.GET(`/`, h.redirect)
	router.GET(`/index`, h.redirect)
	router.GET(`/index.html`, h.redirect)
	router.GET(`/view`, h.redirect)
	router.GET(`/view/`, h.redirect)

	r := router.Group(BaseURL)
	r.GET(`/:locale`, h.viewOrRedirect)
	r.GET(`/:locale/*path`, h.view)
}
func (h Helper) redirect(c *gin.Context) {
}
func (h Helper) viewOrRedirect(c *gin.Context) {
}
func (h Helper) view(c *gin.Context) {
}
