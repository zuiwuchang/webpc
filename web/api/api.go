package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.com/king011/webpc/configure"
	"gitlab.com/king011/webpc/web"
	v1 "gitlab.com/king011/webpc/web/api/v1"
)

// BaseURL request base url
const BaseURL = `api`

// Helper path of /app
type Helper struct {
	web.Helper
}

var maxBytesReader int64

// Register impl IHelper
func (h Helper) Register(router *gin.RouterGroup) {
	maxBytesReader = configure.Single().HTTP.MaxBytesReader
	r := router.Group(BaseURL)
	if maxBytesReader > 0 {
		r.Use(h.CheckBodySize)
	}
	ms := []web.IHelper{
		v1.Helper{},
	}
	for _, m := range ms {
		m.Register(r)
	}
}

// CheckBodySize 限制 body 大小
func (h Helper) CheckBodySize(c *gin.Context) {
	if maxBytesReader > 0 {
		if c.Request != nil && c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytesReader)
		}
	}
}
