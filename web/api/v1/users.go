package v1

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/king011/webpc/web"
)

// Users .
type Users struct {
	web.Helper
}

// Register impl IHelper
func (h Users) Register(router *gin.RouterGroup) {
	r := router.Group(`/users`)
	r.Use(h.CheckRoot)

	r.GET(``, h.list)
	r.POST(``, h.add)
	r.DELETE(`/:id`, h.remove)
	r.PATCH(`/:id/password`, h.password)
	r.PATCH(`/:id/change`, h.change)
}

// list 返回 用戶列表
func (h Users) list(c *gin.Context) {

}

// add 添加 用戶
func (h Users) add(c *gin.Context) {

}

// remove 刪除 用戶
func (h Users) remove(c *gin.Context) {

}

// password 修改 用戶密碼
func (h Users) password(c *gin.Context) {

}

// change 修改 用戶權限
func (h Users) change(c *gin.Context) {

}
