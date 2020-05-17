package daemon

import (
	"github.com/gin-gonic/gin"
)

func newGIN() (router *gin.Engine) {
	router = gin.Default()
	return
}
