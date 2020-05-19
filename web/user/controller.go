package user

import (
	"gitlab.com/king011/webpc/web"

	"github.com/gin-gonic/gin"
)

// BaseURL request base url
const BaseURL = `/user`

// Helper path of /app
type Helper struct {
	web.Helper
}

// Register impl IController
func (h Helper) Register(router *gin.RouterGroup) {
	r := router.Group(BaseURL)

	h.GetPost(r, `/list`, h.list)
	h.GetPost(r, `/add`, h.add)
	h.GetPost(r, `/remove`, h.remove)
	h.GetPost(r, `/password`, h.password)
	h.GetPost(r, `/change`, h.change)
}

// list 返回 用戶列表
func (h Helper) list(c *gin.Context) {

}

// add 添加 用戶
func (h Helper) add(c *gin.Context) {

}

// remove 刪除 用戶
func (h Helper) remove(c *gin.Context) {

}

// password 修改 用戶密碼
func (h Helper) password(c *gin.Context) {

}

// change 修改 用戶權限
func (h Helper) change(c *gin.Context) {

}
