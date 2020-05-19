package app

import (
	"errors"
	"net/http"

	"gitlab.com/king011/webpc/cookie"
	"gitlab.com/king011/webpc/db/manipulator"
	"gitlab.com/king011/webpc/version"
	"gitlab.com/king011/webpc/web"

	"github.com/gin-gonic/gin"
)

// BaseURL request base url
const BaseURL = `/app`

// Helper path of /app
type Helper struct {
	web.Helper
}

// Register impl IController
func (h Helper) Register(router *gin.RouterGroup) {
	r := router.Group(BaseURL)

	h.GetPost(r, `/version`, h.version)
	h.GetPost(r, `/login`, h.login)
	h.GetPost(r, `/restore`, h.restore)
	h.GetPost(r, `/logout`, h.logout)
}

// version 版本信息
func (h Helper) version(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		`tag`:    version.Tag,
		`commit`: version.Commit,
		`date`:   version.Date,
	})
}

// restore 恢復session
func (h Helper) restore(c *gin.Context) {
	session, e := h.GetSession(c)
	if e != nil {
		c.String(http.StatusInternalServerError, e.Error())
		return
	} else if session == nil {
		c.JSON(http.StatusOK, nil)
		return
	}
	c.JSON(http.StatusOK, session)
}

// login 登入
func (h Helper) login(c *gin.Context) {
	var params struct {
		Name     string `form:"name" json:"name" xml:"name" yaml:"name" binding:"required"`
		Password string `form:"password" json:"password" xml:"password" yaml:"password" binding:"required"`
		Remember bool
	}
	e := h.Bind(c, &params)
	if e != nil {
		return
	}
	var mUser manipulator.User
	session, e := mUser.Login(params.Name, params.Password)
	if e != nil {
		c.String(http.StatusInternalServerError, e.Error())
		return
	} else if session == nil {
		e = errors.New("name or password not match")
		c.String(http.StatusInternalServerError, e.Error())
		return
	}
	val, e := session.Cookie()
	if e != nil {
		c.String(http.StatusInternalServerError, e.Error())
		return
	}
	maxage := 0
	if params.Remember {
		maxage = int(cookie.MaxAge())
	}
	c.SetCookie(cookie.CookieName, val, maxage, `/`, ``, false, true)

	c.JSON(http.StatusOK, &session)
}

// logout 登出
func (h Helper) logout(c *gin.Context) {
	c.SetCookie(cookie.CookieName, `expired`, -1, `/`, ``, false, true)
}
