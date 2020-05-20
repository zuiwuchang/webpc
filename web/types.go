package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gitlab.com/king011/webpc/cookie"
)

// Offered accept Offered
var Offered = []string{
	binding.MIMEJSON,
	binding.MIMEHTML,
	binding.MIMEXML,
	binding.MIMEYAML,
}

// IHelper gin 控制器
type IHelper interface {
	// 註冊 控制器
	Register(*gin.RouterGroup)
}

// Helper 爲所有 控制器 定義了 通用的輔助方法
type Helper struct {
}

// Response api 響應
type Response struct {
	Error string      `json:"error,omitempty"`
	Data  interface{} `json:"data,omitempty"`
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
func (h Helper) BindWith(c *gin.Context, obj interface{}, b binding.Binding) (e error) {
	e = c.ShouldBindWith(obj, b)
	if e != nil {
		h.NegotiateError(c, http.StatusBadRequest, e)
		return
	}
	return
}

// CheckRoot 檢查是否具有 root 權限
func (h Helper) CheckRoot(c *gin.Context) {
	session, e := h.GetSession(c)
	if e != nil {
		h.NegotiateError(c, http.StatusUnauthorized, e)
		c.Abort()
		return
	} else if session == nil {
		h.NegotiateErrorString(c, http.StatusUnauthorized, `session miss`)
		c.Abort()
		return
	}
	if !session.Root {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
}

// NegotiateError .
func (h Helper) NegotiateError(c *gin.Context, code int, e error) {
	c.Negotiate(code, gin.Negotiate{
		Offered: Offered,
		Data: Response{
			Error: e.Error(),
		},
	})
}

// NegotiateErrorString .
func (h Helper) NegotiateErrorString(c *gin.Context, code int, e string) {
	c.Negotiate(code, gin.Negotiate{
		Offered: Offered,
		Data: Response{
			Error: e,
		},
	})
}

// NegotiateData .
func (h Helper) NegotiateData(c *gin.Context, code int, data interface{}) {
	c.Negotiate(code, gin.Negotiate{
		Offered: Offered,
		Data: Response{
			Data: data,
		},
	})
}
