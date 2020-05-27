package v1

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

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

	r.GET(``, h.ls)
	r.GET(`:root/:path`, h.get)
	r.PUT(`:root/:path`, h.put)
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
func (h FS) checkWirte(c *gin.Context, m *mount.Mount) (ok bool) {
	session := h.BindSession(c)
	if session == nil {
		return
	}
	if session.Root {
		ok = true
		return
	}
	if !m.Write() {
		c.Status(http.StatusForbidden)
		return
	}

	ok = true
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
