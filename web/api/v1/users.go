package v1

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"gitlab.com/king011/webpc/cookie"
	"gitlab.com/king011/webpc/db/manipulator"
	"gitlab.com/king011/webpc/logger"
	"gitlab.com/king011/webpc/web"
	"go.uber.org/zap"
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
	var mUser manipulator.User
	results, e := mUser.List()
	if e != nil {
		h.NegotiateError(c, http.StatusInternalServerError, e)
		return
	}
	h.NegotiateData(c, http.StatusOK, results)
}

// add 添加 用戶
func (h Users) add(c *gin.Context) {
	session := h.BindSession(c)
	if session == nil {
		if ce := logger.Logger.Check(zap.ErrorLevel, c.FullPath()); ce != nil {
			ce.Write(
				zap.String(`method`, c.Request.Method),
				zap.String(`error`, `session nil`),
			)
		}
		return
	}
	// 解析參數
	var obj struct {
		Name     string `form:"name" json:"name" xml:"name" yaml:"name" binding:"required"`
		Password string `form:"password" json:"password" xml:"password" yaml:"password" binding:"required"`
		Shell    bool   `form:"shell" json:"shell" xml:"shell" yaml:"shell" `
		Read     bool   `form:"read" json:"read" xml:"read" yaml:"read" `
		Write    bool   `form:"write" json:"write" xml:"write" yaml:"write" `
		Root     bool   `form:"root" json:"root" xml:"root" yaml:"root" `
	}
	e := h.Bind(c, &obj)
	if e != nil {
		return
	}
	var mUser manipulator.User
	e = mUser.Add(obj.Name, obj.Password, obj.Shell, obj.Read, obj.Write, obj.Root)
	if e != nil {
		h.NegotiateError(c, http.StatusInternalServerError, e)
		return
	}
	c.Status(http.StatusCreated)

	if ce := logger.Logger.Check(zap.InfoLevel, c.FullPath()); ce != nil {
		s := cookie.Session{
			Name:  obj.Name,
			Shell: obj.Shell,
			Read:  obj.Read,
			Write: obj.Write,
			Root:  obj.Root,
		}
		ce.Write(
			zap.String(`method`, c.Request.Method),
			zap.String(`session`, session.String()),
			zap.String(`new`, s.String()),
			zap.String(`client ip`, c.ClientIP()),
		)
	}
}

// remove 刪除 用戶
func (h Users) remove(c *gin.Context) {
	session := h.BindSession(c)
	if session == nil {
		if ce := logger.Logger.Check(zap.ErrorLevel, c.FullPath()); ce != nil {
			ce.Write(
				zap.String(`method`, c.Request.Method),
				zap.String(`error`, `session nil`),
			)
		}
		return
	}
	var obj struct {
		ID string `uri:"id" binding:"required"`
	}
	e := h.BindURI(c, &obj)
	if e != nil {
		return
	}
	obj.ID, e = url.PathUnescape(obj.ID)
	if e != nil {
		h.NegotiateError(c, http.StatusBadRequest, e)
		return
	}
	var mUser manipulator.User
	e = mUser.Remove(obj.ID)
	if e != nil {
		h.NegotiateError(c, http.StatusInternalServerError, e)
		return
	}
	c.Status(http.StatusNoContent)
	if ce := logger.Logger.Check(zap.WarnLevel, c.FullPath()); ce != nil {
		ce.Write(
			zap.String(`method`, c.Request.Method),
			zap.String(`id`, obj.ID),
		)
	}
	return
}

// password 修改 用戶密碼
func (h Users) password(c *gin.Context) {
	session := h.BindSession(c)
	if session == nil {
		if ce := logger.Logger.Check(zap.ErrorLevel, c.FullPath()); ce != nil {
			ce.Write(
				zap.String(`method`, c.Request.Method),
				zap.String(`error`, `session nil`),
			)
		}
		return
	}

	var obj struct {
		ID string `uri:"id" binding:"required"`
	}
	e := h.BindURI(c, &obj)
	if e != nil {
		return
	}
	obj.ID, e = url.PathUnescape(obj.ID)
	if e != nil {
		h.NegotiateError(c, http.StatusBadRequest, e)
		return
	}
	var val struct {
		Password string `form:"password" json:"password" xml:"password" yaml:"password" binding:"required"`
	}
	e = h.Bind(c, &val)
	if e != nil {
		return
	}

	var mUser manipulator.User
	e = mUser.Password(obj.ID, val.Password)
	if e != nil {
		h.NegotiateError(c, http.StatusInternalServerError, e)
		return
	}
	c.Status(http.StatusNoContent)

	if ce := logger.Logger.Check(zap.WarnLevel, c.FullPath()); ce != nil {
		ce.Write(
			zap.String(`method`, c.Request.Method),
			zap.String(`id`, obj.ID),
		)
	}
}

// change 修改 用戶權限
func (h Users) change(c *gin.Context) {
	session := h.BindSession(c)
	if session == nil {
		if ce := logger.Logger.Check(zap.ErrorLevel, c.FullPath()); ce != nil {
			ce.Write(
				zap.String(`method`, c.Request.Method),
				zap.String(`error`, `session nil`),
			)
		}
		return
	}

	var obj struct {
		ID string `uri:"id" binding:"required"`
	}
	e := h.BindURI(c, &obj)
	if e != nil {
		return
	}
	obj.ID, e = url.PathUnescape(obj.ID)
	if e != nil {
		h.NegotiateError(c, http.StatusBadRequest, e)
		return
	}
	var val struct {
		Shell bool `form:"shell" json:"shell" xml:"shell" yaml:"shell" `
		Read  bool `form:"read" json:"read" xml:"read" yaml:"read" `
		Write bool `form:"write" json:"write" xml:"write" yaml:"write" `
		Root  bool `form:"root" json:"root" xml:"root" yaml:"root" `
	}
	e = h.Bind(c, &val)
	if e != nil {
		return
	}

	var mUser manipulator.User
	e = mUser.Change(obj.ID, val.Shell, val.Read, val.Write, val.Root)
	if e != nil {
		h.NegotiateError(c, http.StatusInternalServerError, e)
		return
	}
	c.Status(http.StatusNoContent)

	if ce := logger.Logger.Check(zap.WarnLevel, c.FullPath()); ce != nil {
		ce.Write(
			zap.String(`method`, c.Request.Method),
			zap.String(`id`, obj.ID),
			zap.Bool(`shell`, val.Shell),
			zap.Bool(`read`, val.Read),
			zap.Bool(`write`, val.Write),
			zap.Bool(`root`, val.Root),
		)
	}
}
