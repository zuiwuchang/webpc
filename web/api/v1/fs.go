package v1

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gitlab.com/king011/webpc/logger"
	"gitlab.com/king011/webpc/mount"
	"gitlab.com/king011/webpc/utils"
	"gitlab.com/king011/webpc/web"
	"go.uber.org/zap"
)

// FS 檔案系統
type FS struct {
	web.Helper
}
type fsURI struct {
	Root string `uri:"root" form:"root" json:"root" xml:"root" yaml:"root" binding:"required"`
	Path string `uri:"path" form:"path" json:"path" xml:"path" yaml:"path" binding:"required"`
}

func (f *fsURI) Unescape() (e error) {
	f.Root, e = url.PathUnescape(f.Root)
	if e != nil {
		return
	}
	f.Path, e = url.PathUnescape(f.Path)
	if e != nil {
		return
	}
	return
}

// Register impl IHelper
func (h FS) Register(router *gin.RouterGroup) {
	r := router.Group(`fs`)
	r.Use(h.CheckSession)

	r.GET(``, h.ls)
	r.GET(`:root/:path`, h.get)
	r.PUT(`:root/:path`, h.CheckSession, h.put)
	r.PATCH(`:root/:path/name`, h.CheckSession, h.rename)
	r.POST(`:root/:path`, h.CheckSession, h.post)
	r.DELETE(`:root/:path`, h.CheckSession, h.remove)
	r.GET(`:root/:path/compress/websocket`, h.CheckWebsocket, h.CheckSession, h.compress)
}
func (h FS) ls(c *gin.Context) {
	var obj struct {
		Root string `form:"root" json:"root" xml:"root" yaml:"root" binding:"required"`
		Path string `form:"path" json:"path" xml:"path" yaml:"path" binding:"required"`
	}
	e := h.BindQuery(c, &obj)
	if e != nil {
		return
	}
	fs := mount.Single()
	m := fs.Root(obj.Root)
	if m == nil {
		c.Status(http.StatusNotFound)
		return
	}
	if !h.checkRead(c, m) {
		return
	}
	dir, items, e := m.LS(obj.Path)
	if e != nil {
		h.NegotiateError(c, http.StatusForbidden, e)
		return
	}

	h.NegotiateData(c, http.StatusOK, gin.H{
		`dir`: gin.H{
			`root`:   m.Name(),
			`read`:   m.Read(),
			`write`:  m.Write(),
			`shared`: m.Shared(),
			`dir`:    dir,
		},
		`items`: items,
	})
}
func (h FS) checkRead(c *gin.Context, m *mount.Mount) (ok bool) {
	if m.Shared() {
		ok = true
		return
	}
	session := h.BindSession(c)
	if session == nil {
		return
	}
	if session.Root {
		ok = true
		return
	}
	if !m.Read() {
		c.Status(http.StatusForbidden)
		return
	}

	ok = true
	return
}
func (h FS) canWirte(c *gin.Context, m *mount.Mount) (ok bool) {
	session := h.BindSession(c)
	if session == nil {
		return
	}
	if session.Root {
		ok = true
		return
	}
	if !m.Write() {
		return
	}
	ok = true
	return
}
func (h FS) checkWirte(c *gin.Context, m *mount.Mount) (ok bool) {
	if !h.canWirte(c, m) {
		c.Status(http.StatusForbidden)
		return
	}
	ok = true
	return
}
func (h FS) bindURINormal(c *gin.Context) (obj fsURI, e error) {
	e = c.ShouldBindUri(&obj)
	if e != nil {
		return
	}
	e = obj.Unescape()
	if e != nil {
		return
	}
	return
}
func (h FS) bindURI(c *gin.Context) (obj fsURI, e error) {
	e = h.BindURI(c, &obj)
	if e != nil {
		return
	}
	e = obj.Unescape()
	if e != nil {
		h.NegotiateError(c, http.StatusBadRequest, e)
		return
	}
	return
}
func (h FS) get(c *gin.Context) {
	obj, e := h.bindURI(c)
	if e != nil {
		return
	}

	fs := mount.Single()
	m := fs.Root(obj.Root)
	if m == nil {
		c.Status(http.StatusNotFound)
		return
	}
	if !h.checkRead(c, m) {
		return
	}
	filename, e := m.Filename(obj.Path)
	if e != nil {
		h.NegotiateError(c, http.StatusForbidden, e)
		return
	}
	c.FileAttachment(filename, filepath.Base(filename))
}
func (h FS) put(c *gin.Context) {
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
	obj, e := h.bindURI(c)
	if e != nil {
		return
	}
	fs := mount.Single()
	m := fs.Root(obj.Root)
	if m == nil {
		c.Status(http.StatusNotFound)
		return
	}
	if !h.checkWirte(c, m) {
		return
	}
	filename, e := m.Filename(obj.Path)
	if e != nil {
		h.NegotiateError(c, http.StatusForbidden, e)
		return
	}

	var b []byte
	if c.ContentType() == binding.MIMEPlain {
		req := c.Request
		if req != nil && req.Body != nil {
			b, e = ioutil.ReadAll(req.Body)
			if e != nil {
				h.NegotiateError(c, http.StatusBadRequest, e)
				return
			}
		}
	} else {
		var obj struct {
			Val string `form:"val" json:"val" xml:"val" yaml:"val" binding:"required"`
		}
		e = h.Bind(c, &obj)
		if e != nil {
			return
		}
		if obj.Val != "" {
			b = utils.StringToBytes(obj.Val)
		}
	}
	e = h.writeFile(filename, b, 0666)
	if e != nil {
		h.NegotiateError(c, http.StatusForbidden, e)
		return
	}

	if ce := logger.Logger.Check(zap.WarnLevel, c.FullPath()); ce != nil {
		ce.Write(
			zap.String(`method`, c.Request.Method),
			zap.String(`session`, session.String()),
			zap.String(`root`, obj.Root),
			zap.String(`path`, obj.Path),
		)
	}
	c.Status(http.StatusNoContent)
}
func (h FS) writeFile(filename string, data []byte, perm os.FileMode) error {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	n, err := f.Write(data)
	if err == nil && n < len(data) {
		err = io.ErrShortWrite
	}
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return err
}
func (h FS) rename(c *gin.Context) {
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
	objURI, e := h.bindURI(c)
	if e != nil {
		return
	}
	fs := mount.Single()
	m := fs.Root(objURI.Root)
	if m == nil {
		c.Status(http.StatusNotFound)
		return
	}
	if !h.checkWirte(c, m) {
		return
	}
	filename, e := m.Filename(objURI.Path)
	if e != nil {
		h.NegotiateError(c, http.StatusForbidden, e)
		return
	}

	var obj struct {
		Val string `form:"val" json:"val" xml:"val" yaml:"val" binding:"required"`
	}
	e = h.Bind(c, &obj)
	if e != nil {
		return
	}
	dst := filepath.Base(filepath.Clean(obj.Val))
	if dst != obj.Val {
		h.NegotiateErrorString(c, http.StatusBadRequest, `name not support`)
		return
	}
	dst = filepath.Dir(filename) + `/` + dst
	e = os.Rename(filename, dst)
	if e != nil {
		h.NegotiateError(c, http.StatusForbidden, e)
		return
	}
	if ce := logger.Logger.Check(zap.WarnLevel, c.FullPath()); ce != nil {
		ce.Write(
			zap.String(`method`, c.Request.Method),
			zap.String(`session`, session.String()),
			zap.String(`root`, objURI.Root),
			zap.String(`path`, objURI.Path),
			zap.String(`val`, obj.Val),
			zap.String(`src`, filename),
			zap.String(`dst`, dst),
		)
	}
	c.Status(http.StatusNoContent)
}
func (h FS) post(c *gin.Context) {
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
	objURI, e := h.bindURI(c)
	if e != nil {
		return
	}
	fs := mount.Single()
	m := fs.Root(objURI.Root)
	if m == nil {
		c.Status(http.StatusNotFound)
		return
	}
	if !h.checkWirte(c, m) {
		return
	}
	filename, e := m.Filename(objURI.Path)
	if e != nil {
		h.NegotiateError(c, http.StatusForbidden, e)
		return
	}
	var obj struct {
		Dir  bool   `form:"dir" json:"dir" xml:"dir" yaml:"dir"`
		Name string `form:"name" json:"name" xml:"name" yaml:"name" binding:"required"`
	}
	e = h.Bind(c, &obj)
	if e != nil {
		return
	}
	dst := filepath.Base(filepath.Clean(obj.Name))
	if dst != obj.Name {
		h.NegotiateErrorString(c, http.StatusBadRequest, `name not support`)
		return
	}
	dst = filename + `/` + dst
	var f *os.File
	var info mount.FileInfo
	info.Name = obj.Name
	if obj.Dir {
		e = os.Mkdir(dst, 0775)
		if e != nil {
			h.NegotiateError(c, http.StatusForbidden, e)
			return
		}
		info.IsDir = true
	} else {
		f, e = os.OpenFile(dst, os.O_CREATE|os.O_EXCL, 0666)
		if e != nil {
			h.NegotiateError(c, http.StatusForbidden, e)
			return
		}
		f.Close()
	}
	stat, _ := os.Stat(dst)
	if stat != nil {
		info.IsDir = stat.IsDir()
		info.Size = stat.Size()
		info.Mode = uint32(stat.Mode())
	}
	if ce := logger.Logger.Check(zap.WarnLevel, c.FullPath()); ce != nil {
		ce.Write(
			zap.String(`method`, c.Request.Method),
			zap.String(`session`, session.String()),
			zap.String(`root`, objURI.Root),
			zap.String(`dir`, objURI.Path),
			zap.String(`name`, obj.Name),
			zap.Bool(`dir`, obj.Dir),
			zap.String(`dst`, dst),
		)
	}
	h.NegotiateData(c, http.StatusCreated, &info)
}
func (h FS) remove(c *gin.Context) {
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
	objURI, e := h.bindURI(c)
	if e != nil {
		return
	}
	fs := mount.Single()
	m := fs.Root(objURI.Root)
	if m == nil {
		c.Status(http.StatusNotFound)
		return
	}
	if !h.checkWirte(c, m) {
		return
	}
	filename, e := m.Filename(objURI.Path)
	if e != nil {
		h.NegotiateError(c, http.StatusForbidden, e)
		return
	}

	var obj struct {
		Names []string `form:"names" json:"names" xml:"names" yaml:"names" binding:"required"`
	}
	e = h.Bind(c, &obj)
	if e != nil {
		return
	}
	count := len(obj.Names)
	if count == 0 {
		h.NegotiateErrorString(c, http.StatusBadRequest, `names nil`)
		return
	}
	dsts := make([]string, count)
	for i := 0; i < count; i++ {
		dst := filepath.Base(filepath.Clean(obj.Names[i]))
		if dst != obj.Names[i] {
			h.NegotiateErrorString(c, http.StatusBadRequest, `name not support`)
			return
		}
		dsts[i] = filepath.Clean(filename + `/` + dst)
	}
	for i := 0; i < count; i++ {
		e := os.RemoveAll(dsts[i])
		if e != nil {
			if i != 0 {
				if ce := logger.Logger.Check(zap.ErrorLevel, c.FullPath()); ce != nil {
					ce.Write(
						zap.String(`method`, c.Request.Method),
						zap.Error(e),
						zap.String(`root`, objURI.Root),
						zap.String(`dir`, objURI.Path),
						zap.Strings(`names`, obj.Names),
						zap.Strings(`dsts`, dsts[:i]),
					)
				}
			}
			h.NegotiateError(c, http.StatusInternalServerError, e)
			return
		}
	}
	if ce := logger.Logger.Check(zap.WarnLevel, c.FullPath()); ce != nil {
		ce.Write(
			zap.String(`method`, c.Request.Method),
			zap.String(`session`, session.String()),
			zap.String(`root`, objURI.Root),
			zap.String(`dir`, objURI.Path),
			zap.Strings(`names`, obj.Names),
			zap.Strings(`dsts`, dsts),
		)
	}
	c.Status(http.StatusNoContent)
}
func (h FS) compress(c *gin.Context) {
	ws, e := upgrader.Upgrade(c.Writer, c.Request, nil)
	if e != nil {
		h.NegotiateError(c, http.StatusBadRequest, e)
		return
	}
	defer ws.Close()
	session := h.BindSession(c)
	if session == nil {
		if ce := logger.Logger.Check(zap.ErrorLevel, c.FullPath()); ce != nil {
			ce.Write(
				zap.String(`method`, c.Request.Method),
				zap.String(`error`, `session nil`),
			)
		}
		h.WriteJSON(ws, gin.H{
			`cmd`:   mount.CmdError,
			`error`: `session nil`,
		})
		return
	}
	objURI, e := h.bindURINormal(c)
	if e != nil {
		h.WriteJSON(ws, gin.H{
			`cmd`:   mount.CmdError,
			`error`: e.Error(),
		})
		return
	}
	fs := mount.Single()
	m := fs.Root(objURI.Root)
	if m == nil {
		h.WriteJSON(ws, gin.H{
			`cmd`:   mount.CmdError,
			`error`: `not found`,
		})
		return
	}
	if !h.canWirte(c, m) {
		h.WriteJSON(ws, gin.H{
			`cmd`:   mount.CmdError,
			`error`: `Forbidden`,
		})
		return
	}
	dir, e := m.Filename(objURI.Path)
	if e != nil {
		h.WriteJSON(ws, gin.H{
			`cmd`:   mount.CmdError,
			`error`: e.Error(),
		})
		return
	}
	name, names, e := mount.Compress(ws, dir, time.Second*10)
	if e != nil {
		if ce := logger.Logger.Check(zap.WarnLevel, c.FullPath()); ce != nil {
			ce.Write(
				zap.Error(e),
				zap.String(`method`, c.Request.Method),
				zap.String(`session`, session.String()),
				zap.String(`root`, objURI.Root),
				zap.String(`dir`, objURI.Path),
				zap.Strings(`names`, names),
				zap.String(`name`, name),
			)
		}
		return
	}
	if ce := logger.Logger.Check(zap.WarnLevel, c.FullPath()); ce != nil {
		ce.Write(
			zap.String(`method`, c.Request.Method),
			zap.String(`session`, session.String()),
			zap.String(`root`, objURI.Root),
			zap.String(`dir`, objURI.Path),
			zap.Strings(`names`, names),
			zap.String(`name`, name),
		)
	}
}
