package daemon

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/king011/webpc/web"
	"gitlab.com/king011/webpc/web/api"
	"gitlab.com/king011/webpc/web/view"
)

func newGIN() (router *gin.Engine) {
	router = gin.Default()
	rs := []web.IHelper{
		view.Helper{},
		api.Helper{},
	}
	for _, r := range rs {
		r.Register(&router.RouterGroup)
	}
	return
}
