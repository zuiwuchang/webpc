package daemon

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/king011/webpc/web"
	"gitlab.com/king011/webpc/web/app"
)

func newGIN() (router *gin.Engine) {
	router = gin.Default()
	rs := []web.IHelper{
		app.Helper{},
	}
	for _, r := range rs {
		r.Register(&router.RouterGroup)
	}
	return
}
