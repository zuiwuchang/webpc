package daemon

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/king011/webpc/web"
	"gitlab.com/king011/webpc/web/app"
	"gitlab.com/king011/webpc/web/user"
)

func newGIN() (router *gin.Engine) {
	router = gin.Default()
	rs := []web.IHelper{
		app.Helper{},
		user.Helper{},
	}
	for _, r := range rs {
		r.Register(&router.RouterGroup)
	}
	return
}
