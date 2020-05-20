package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.com/king011/webpc/cookie"
	"gitlab.com/king011/webpc/db/manipulator"
	"gitlab.com/king011/webpc/web"
)

// Session .
type Session struct {
	web.Helper
}

// Register impl IHelper
func (h Session) Register(router *gin.RouterGroup) {
	router.PUT(`/session`, h.login)
	router.GET(`/session`, h.restore)
	router.DELETE(`/session`, h.logout)
}
func (h Session) login(c *gin.Context) {
	// 解析參數
	var obj struct {
		Name     string `form:"name" json:"name" xml:"name" yaml:"name" binding:"required"`
		Password string `form:"password" json:"password" xml:"password" yaml:"password" binding:"required"`
		Remember bool   `form:"remember" json:"remember" xml:"remember" yaml:"remember" `
	}
	e := h.Bind(c, &obj)
	if e != nil {
		return
	}

	// 查詢用戶
	var mUser manipulator.User
	session, e := mUser.Login(obj.Name, obj.Password)
	if e != nil {
		h.NegotiateError(c, http.StatusInternalServerError, e)
		return
	} else if session == nil {
		h.NegotiateErrorString(c, http.StatusNotFound, `name or password not match`)
		return
	}
	// 生成 cookie
	val, e := session.Cookie()
	if e != nil {
		h.NegotiateError(c, http.StatusInternalServerError, e)
		return
	}
	maxage := 0
	if obj.Remember {
		maxage = int(cookie.MaxAge())
	}

	// 響應 數據
	c.SetCookie(cookie.CookieName, val, maxage, `/`, ``, false, true)
	h.NegotiateData(c, http.StatusCreated, session)
}
func (h Session) restore(c *gin.Context) {
	session, e := h.GetSession(c)
	if e != nil {
		h.NegotiateError(c, http.StatusUnauthorized, e)
		return
	} else if session == nil {
		h.NegotiateErrorString(c, http.StatusNotFound, `not found`)
		return
	}
	h.NegotiateData(c, http.StatusOK, session)
}
func (h Session) logout(c *gin.Context) {
	c.SetCookie(cookie.CookieName, `expired`, -1, `/`, ``, false, true)
	c.Status(http.StatusNoContent)
}
