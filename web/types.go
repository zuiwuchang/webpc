package web

import (
	"fmt"
	"net/http"
	"reflect"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gitlab.com/king011/webpc/cookie"
	"gitlab.com/king011/webpc/logger"
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

// ShouldBindSession 返回session 不進行響應
func (Helper) ShouldBindSession(c *gin.Context) (session *cookie.Session, e error) {
	v, exists := c.Get(`session`)
	if exists {
		if v == nil {
			return
		} else if tmp, ok := v.(error); ok {
			e = tmp
			return
		} else if tmp, ok := v.(*cookie.Session); ok {
			session = tmp
			return
		}
		if ce := logger.Logger.Check(zap.ErrorLevel, `unknow session type`); ce != nil {
			ce.Write(
				zap.String(`method`, c.Request.Method),
				zap.String(`session`, fmt.Sprint(session)),
				zap.String(`session type`, fmt.Sprint(reflect.TypeOf(session))),
			)
		}
		return
	}
	val, e := c.Cookie(cookie.CookieName)
	if e != nil {
		c.Set(`session`, e)
		return
	} else if val == "" {
		c.Set(`session`, nil)
		return
	}
	session, e = cookie.FromCookie(val)
	if e == nil {
		c.Set(`session`, session)
	} else {
		c.Set(`session`, e)
	}
	return
}

// BindSession 返回 session 並響應錯誤
func (h Helper) BindSession(c *gin.Context) (result *cookie.Session) {
	session, e := h.ShouldBindSession(c)
	if e != nil {
		h.NegotiateError(c, http.StatusUnauthorized, e)
		return
	} else if session == nil {
		h.NegotiateErrorString(c, http.StatusUnauthorized, `session miss`)
		return
	}
	result = session
	return
}

// Bind .
func (h Helper) Bind(c *gin.Context, obj interface{}) error {
	b := binding.Default(c.Request.Method, c.ContentType())
	return h.BindWith(c, obj, b)
}

// BindURI .
func (h Helper) BindURI(c *gin.Context, obj interface{}) (e error) {
	e = c.ShouldBindUri(obj)
	if e != nil {
		h.NegotiateError(c, http.StatusBadRequest, e)
		return
	}
	return
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
	session := h.BindSession(c)
	if session == nil {
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
		Data:    e.Error(),
	})
}

// NegotiateErrorString .
func (h Helper) NegotiateErrorString(c *gin.Context, code int, e string) {
	c.Negotiate(code, gin.Negotiate{
		Offered: Offered,
		Data:    e,
	})
}

// NegotiateData .
func (h Helper) NegotiateData(c *gin.Context, code int, data interface{}) {
	c.Negotiate(code, gin.Negotiate{
		Offered: Offered,
		Data:    data,
	})
}
