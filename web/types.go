package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gitlab.com/king011/webpc/cookie"
)

// IHelper gin 控制器
type IHelper interface {
	// 註冊 控制器
	Register(*gin.RouterGroup)
}

// Helper 爲所有 控制器 定義了 通用的輔助方法
type Helper struct {
}

// GetPost r.Get r.POST
func (Helper) GetPost(r *gin.RouterGroup, relativePath string, handlers ...gin.HandlerFunc) {
	r.GET(relativePath, handlers...)
	r.POST(relativePath, handlers...)
}

// GetSession 返回 session
func (Helper) GetSession(c *gin.Context) (session *cookie.Session, e error) {
	val, e := c.Cookie(cookie.CookieName)
	if e != nil {
		return
	} else if val == "" {
		return
	}
	session, e = cookie.FromCookie(val)
	return
}

// Bind .
func (h Helper) Bind(c *gin.Context, obj interface{}) error {
	b := binding.Default(c.Request.Method, c.ContentType())
	return h.BindWith(c, obj, b)
}

// BindJSON .
func (h Helper) BindJSON(c *gin.Context, obj interface{}) error {
	return h.BindWith(c, obj, binding.JSON)
}

// BindWith .
func (Helper) BindWith(c *gin.Context, obj interface{}, b binding.Binding) (e error) {
	e = c.ShouldBindWith(obj, b)
	if e != nil {
		c.String(http.StatusBadRequest, e.Error())
		return
	}
	return
}
