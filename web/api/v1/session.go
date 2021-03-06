package v1

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
	"gitlab.com/king011/webpc/cookie"
	"gitlab.com/king011/webpc/db/manipulator"
	"gitlab.com/king011/webpc/logger"
	"gitlab.com/king011/webpc/web"
)

var _sessionLock _SessionLock
var errSessionLock = errors.New(`lock session error`)
var errIPLock = errors.New(`lock ip error`)

func init() {
	_sessionLock.keys = make(map[string]int8)
	_sessionLock.keysIP = make(map[string]int8)
}

// 防止暴力猜密碼
type _SessionLock struct {
	keys   map[string]int8
	keysIP map[string]int8
	mutex  sync.Mutex
}

func (s *_SessionLock) Lock(key string) (e error) {
	s.mutex.Lock()
	num := s.keys[key]
	if num < 2 {
		num++
		s.keys[key] = num
	} else {
		e = errSessionLock
	}
	s.mutex.Unlock()
	if e == nil {
		time.Sleep(time.Second)
	}
	return
}
func (s *_SessionLock) Unlock(key string) {
	s.mutex.Lock()
	if num, ok := s.keys[key]; ok {
		if num < 2 {
			if len(s.keys) > 30 {
				delete(s.keys, key)
			} else {
				s.keys[key] = 0
			}
		} else {
			s.keys[key] = num - 1
		}
	}
	s.mutex.Unlock()
}
func (s *_SessionLock) LockIP(key string) (e error) {
	s.mutex.Lock()
	num := s.keysIP[key]
	if num < 90 {
		num++
		s.keysIP[key] = num
	} else {
		e = errIPLock
	}
	s.mutex.Unlock()
	return
}
func (s *_SessionLock) UnlockIP(key string) {
	s.mutex.Lock()
	if num, ok := s.keysIP[key]; ok {
		if num < 2 {
			if len(s.keysIP) > 90 {
				delete(s.keysIP, key)
			} else {
				s.keysIP[key] = 0
			}
		} else {
			s.keysIP[key] = num - 1
		}
	}
	s.mutex.Unlock()
}

// Session .
type Session struct {
	web.Helper
}

// Register impl IHelper
func (h Session) Register(router *gin.RouterGroup) {
	r := router.Group(`/session`)
	r.POST(``, h.login)
	r.GET(``, h.restore)
	r.DELETE(``, h.logout)
	r.PATCH(`/password`, h.password)
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
	ip := c.ClientIP()
	e = _sessionLock.LockIP(ip)
	if e != nil {
		h.NegotiateError(c, http.StatusForbidden, e)
		return
	}
	defer _sessionLock.UnlockIP(ip)
	key := ip + obj.Name
	e = _sessionLock.Lock(key)
	if e != nil {
		h.NegotiateError(c, http.StatusForbidden, e)
		return
	}
	defer _sessionLock.Unlock(key)

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

	if ce := logger.Logger.Check(zap.WarnLevel, c.FullPath()); ce != nil {
		ce.Write(
			zap.String(`method`, c.Request.Method),
			zap.String(`session`, session.String()),
			zap.Int(`maxage`, maxage),
			zap.String(`client ip`, c.ClientIP()),
		)
	}
}
func (h Session) restore(c *gin.Context) {
	session, e := h.ShouldBindSession(c)
	if e != nil {
		h.NegotiateError(c, http.StatusUnauthorized, e)
		return
	} else if session == nil {
		h.NegotiateErrorString(c, http.StatusNotFound, `not found`)
		return
	}
	h.NegotiateData(c, http.StatusOK, session)

	if ce := logger.Logger.Check(zap.InfoLevel, c.FullPath()); ce != nil {
		ce.Write(
			zap.String(`method`, c.Request.Method),
			zap.String(`session`, session.String()),
			zap.String(`client ip`, c.ClientIP()),
		)
	}
}
func (h Session) logout(c *gin.Context) {
	c.SetCookie(cookie.CookieName, `expired`, -1, `/`, ``, false, true)
	c.Status(http.StatusNoContent)
}
func (h Session) password(c *gin.Context) {
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
		Old string `form:"old" json:"old" xml:"old" yaml:"old" binding:"required"`
		Val string `form:"val" json:"val" xml:"val" yaml:"val" binding:"required"`
	}
	e := h.Bind(c, &obj)
	if e != nil {
		return
	}
	if obj.Old == obj.Val {
		h.NegotiateErrorString(c, http.StatusBadGateway, `password not changed`)
		return
	}
	var mUser manipulator.User
	e = mUser.Password3(session.Name, obj.Old, obj.Val)
	if e != nil {
		h.NegotiateError(c, http.StatusInternalServerError, e)
		return
	}

	c.Status(http.StatusNoContent)
	if ce := logger.Logger.Check(zap.InfoLevel, c.FullPath()); ce != nil {
		ce.Write(
			zap.String(`method`, c.Request.Method),
			zap.String(`session`, session.String()),
			zap.String(`client ip`, c.ClientIP()),
		)
	}
}
