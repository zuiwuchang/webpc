package v1

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"math"
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
type fsURI2 struct {
	Root    string `uri:"root" form:"root" json:"root" xml:"root" yaml:"root" binding:"required"`
	Path    string `uri:"path" form:"path" json:"path" xml:"path" yaml:"path" binding:"required"`
	SrcRoot string `uri:"srcroot" form:"srcroot" json:"srcroot" xml:"srcroot" yaml:"srcroot" binding:"required"`
	SrcPath string `uri:"srcpath" form:"srcpath" json:"srcpath" xml:"srcpath" yaml:"srcpath" binding:"required"`
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
func (f *fsURI2) Unescape() (e error) {
	f.Root, e = url.PathUnescape(f.Root)
	if e != nil {
		return
	}
	f.Path, e = url.PathUnescape(f.Path)
	if e != nil {
		return
	}
	f.SrcRoot, e = url.PathUnescape(f.SrcRoot)
	if e != nil {
		return
	}
	f.SrcPath, e = url.PathUnescape(f.SrcPath)
	if e != nil {
		return
	}
	return
}

type fsChunkURI struct {
	fsURI
	Index int `uri:"index" form:"index" json:"index" xml:"index" yaml:"index"`
}

// Register impl IHelper
func (h FS) Register(router *gin.RouterGroup) {
	r := router.Group(`fs`)

	r.GET(``, h.ls)
	r.GET(`:root/:path`, h.CheckSession, h.get)
	r.PUT(`:root/:path`, h.CheckSession, h.put)
	r.PATCH(`:root/:path/name`, h.CheckSession, h.rename)
	r.POST(`:root/:path`, h.CheckSession, h.post)
	r.DELETE(`:root/:path`, h.CheckSession, h.remove)
	r.GET(`:root/:path/compress/websocket`, h.CheckWebsocket, h.CheckSession, h.compress)
	r.GET(`:root/:path/uncompress/websocket`, h.CheckWebsocket, h.CheckSession, h.uncompress)
	r.GET(`:root/:path/cut/:srcroot/:srcpath/websocket`, h.CheckWebsocket, h.CheckSession, h.cut)
	r.GET(`:root/:path/copy/:srcroot/:srcpath/websocket`, h.CheckWebsocket, h.CheckSession, h.copy)
	r.GET(`:root/:path/whash`, h.CheckSession, h.whash)
	r.GET(`:root/:path/wchunk`, h.CheckSession, h.wchunk)
	r.PUT(`:root/:path/wchunk/:index`, h.CheckSession, h.putChunk)
	r.PUT(`:root/:path/merge`, h.CheckSession, h.merge)

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
	if !h.canRead(c, m) {
		c.Status(http.StatusForbidden)
		return
	}
	ok = true
	return
}
func (h FS) canRead(c *gin.Context, m *mount.Mount) (ok bool) {
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
func (h FS) bind2URINormal(c *gin.Context) (obj fsURI2, e error) {
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
func (h FS) bindChunkURI(c *gin.Context) (obj fsChunkURI, e error) {
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
	session, _ := h.ShouldBindSession(c)
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
func (h FS) uncompress(c *gin.Context) {
	ws, e := upgrader.Upgrade(c.Writer, c.Request, nil)
	if e != nil {
		h.NegotiateError(c, http.StatusBadRequest, e)
		return
	}
	defer ws.Close()
	session, _ := h.ShouldBindSession(c)
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
	name, e := mount.Uncompress(ws, dir, time.Second*10)
	if e != nil {
		if ce := logger.Logger.Check(zap.WarnLevel, c.FullPath()); ce != nil {
			ce.Write(
				zap.Error(e),
				zap.String(`method`, c.Request.Method),
				zap.String(`session`, session.String()),
				zap.String(`root`, objURI.Root),
				zap.String(`dir`, objURI.Path),
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
			zap.String(`name`, name),
		)
	}
}
func (h FS) cut(c *gin.Context) {
	ws, e := upgrader.Upgrade(c.Writer, c.Request, nil)
	if e != nil {
		h.NegotiateError(c, http.StatusBadRequest, e)
		return
	}
	defer ws.Close()
	session, _ := h.ShouldBindSession(c)
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
	objURI, e := h.bind2URINormal(c)
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

	srcM := fs.Root(objURI.SrcRoot)
	if srcM == nil {
		h.WriteJSON(ws, gin.H{
			`cmd`:   mount.CmdError,
			`error`: `not found`,
		})
		return
	}
	if !h.canWirte(c, srcM) {
		h.WriteJSON(ws, gin.H{
			`cmd`:   mount.CmdError,
			`error`: `Forbidden`,
		})
		return
	}
	srcDir, e := srcM.Filename(objURI.SrcPath)
	if e != nil {
		h.WriteJSON(ws, gin.H{
			`cmd`:   mount.CmdError,
			`error`: e.Error(),
		})
		return
	}

	names, e := mount.Cut(ws, dir, srcDir, time.Second*10)
	if e != nil {
		if ce := logger.Logger.Check(zap.WarnLevel, c.FullPath()); ce != nil {
			ce.Write(
				zap.Error(e),
				zap.String(`method`, c.Request.Method),
				zap.String(`session`, session.String()),
				zap.String(`root`, objURI.Root),
				zap.String(`dir`, objURI.Path),
				zap.String(`src root`, objURI.SrcRoot),
				zap.String(`src dir`, objURI.SrcPath),
				zap.Strings(`names`, names),
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
			zap.String(`src root`, objURI.SrcPath),
			zap.String(`src dir`, objURI.SrcPath),
			zap.Strings(`names`, names),
		)
	}
}
func (h FS) copy(c *gin.Context) {
	ws, e := upgrader.Upgrade(c.Writer, c.Request, nil)
	if e != nil {
		h.NegotiateError(c, http.StatusBadRequest, e)
		return
	}
	defer ws.Close()
	session, _ := h.ShouldBindSession(c)
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
	objURI, e := h.bind2URINormal(c)
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

	srcM := fs.Root(objURI.SrcRoot)
	if srcM == nil {
		h.WriteJSON(ws, gin.H{
			`cmd`:   mount.CmdError,
			`error`: `not found`,
		})
		return
	}
	if !h.canRead(c, srcM) {
		h.WriteJSON(ws, gin.H{
			`cmd`:   mount.CmdError,
			`error`: `Forbidden`,
		})
		return
	}
	srcDir, e := srcM.Filename(objURI.SrcPath)
	if e != nil {
		h.WriteJSON(ws, gin.H{
			`cmd`:   mount.CmdError,
			`error`: e.Error(),
		})
		return
	}

	names, e := mount.Copy(ws, dir, srcDir, time.Second*10)
	if e != nil {
		if ce := logger.Logger.Check(zap.WarnLevel, c.FullPath()); ce != nil {
			ce.Write(
				zap.Error(e),
				zap.String(`method`, c.Request.Method),
				zap.String(`session`, session.String()),
				zap.String(`root`, objURI.Root),
				zap.String(`dir`, objURI.Path),
				zap.String(`src root`, objURI.SrcRoot),
				zap.String(`src dir`, objURI.SrcPath),
				zap.Strings(`names`, names),
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
			zap.String(`src root`, objURI.SrcPath),
			zap.String(`src dir`, objURI.SrcPath),
			zap.Strings(`names`, names),
		)
	}
}

func (h FS) whash(c *gin.Context) {
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
	var params struct {
		Chunk int64 `uri:"chunk" form:"chunk" json:"chunk" xml:"chunk" yaml:"chunk" binding:"required"`
		Size  int64 `uri:"size" form:"size" json:"size" xml:"size" yaml:"size" binding:"required"`
	}
	e = h.BindQuery(c, &params)
	if e != nil {
		return
	}
	if params.Size < 1 {
		h.NegotiateErrorString(c, http.StatusBadRequest, `not support size`)
		return
	}
	if params.Chunk < 1024*1024 && params.Chunk > 1024*1024*50 {
		h.NegotiateErrorString(c, http.StatusBadRequest, `not support chunk size`)
		return
	}

	f, e := os.Open(filename)
	if e != nil {
		if os.IsNotExist(e) {
			c.Status(http.StatusNoContent)
			return
		}
		h.NegotiateError(c, http.StatusForbidden, e)
		return
	}
	defer f.Close()
	size, e := f.Seek(0, os.SEEK_END)
	if e != nil {
		h.NegotiateError(c, http.StatusForbidden, e)
		return
	}
	f.Seek(0, os.SEEK_SET)
	if e != nil {
		h.NegotiateError(c, http.StatusForbidden, e)
		return
	}
	if size != params.Size {
		h.NegotiateData(c, http.StatusOK, `size not match`)
		return
	}
	hash := md5.New()
	b := make([]byte, params.Chunk)
	var n int
	buffer := make([]byte, 32)
	for {
		n, e = f.Read(b)
		if n != 0 {
			val := md5.Sum(b[:n])
			hex.Encode(buffer, val[:])
			_, e = hash.Write(buffer)
			if e != nil {
				h.NegotiateError(c, http.StatusInternalServerError, e)
				return
			}
		}
		if e == io.EOF {
			break
		}
		if e != nil {
			h.NegotiateError(c, http.StatusInternalServerError, e)
			return
		}
	}
	h.NegotiateData(c, http.StatusOK, hex.EncodeToString(hash.Sum(nil)))
}
func (h FS) wchunk(c *gin.Context) {
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
	var params struct {
		Start int `uri:"start" form:"start" json:"start" xml:"start" yaml:"start" `
		Count int `uri:"count" form:"count" json:"count" xml:"count" yaml:"count" binding:"required"`
	}
	e = h.BindQuery(c, &params)
	if e != nil {
		return
	}
	if params.Start < 0 || params.Start > math.MaxInt32-1000 {
		h.NegotiateErrorString(c, http.StatusBadRequest, `not support start`)
		return
	}
	if params.Count < 1 || params.Count > 1000 {
		h.NegotiateErrorString(c, http.StatusBadRequest, `not support count`)
		return
	}
	dir, name := filepath.Split(filename)
	dir = filepath.Clean(dir + `/.chunks_` + name)
	results := make([]string, params.Count)
	var f *os.File
	chunk := md5.New()
	for i := 0; i < params.Count; i++ {
		filename := dir + `/` + fmt.Sprint(params.Start+i)
		f, e = os.Open(filename)
		if e != nil {
			if os.IsNotExist(e) {
				continue
			}
			h.NegotiateError(c, http.StatusInternalServerError, e)
			return
		}
		chunk.Reset()
		_, e = io.Copy(chunk, f)
		f.Close()
		if e != nil {
			h.NegotiateError(c, http.StatusInternalServerError, e)
			return
		}
		val := chunk.Sum(nil)
		results[i] = hex.EncodeToString(val[:])
	}
	h.NegotiateData(c, http.StatusOK, results)
}
func (h FS) putChunk(c *gin.Context) {
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
	objURI, e := h.bindChunkURI(c)
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
	if c.Request.Body == nil {
		h.NegotiateErrorString(c, http.StatusBadRequest, `body nil`)
		return
	}
	dir, name := filepath.Split(filename)
	dir = filepath.Clean(dir + `/.chunks_` + name)
	os.Mkdir(dir, 0775)
	f, e := os.Create(dir + `/` + fmt.Sprint(objURI.Index))
	if e != nil {
		h.NegotiateError(c, http.StatusForbidden, e)
		return
	}
	_, e = io.Copy(f, c.Request.Body)
	if e != nil {
		h.NegotiateError(c, http.StatusForbidden, e)
		return
	}
	c.Request.Body.Close()
	c.Status(http.StatusCreated)
}
func (h FS) merge(c *gin.Context) {
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
	var params struct {
		Hash  string `uri:"hash" form:"hash" json:"hash" xml:"hash" yaml:"hash" binding:"required"`
		Count int    `uri:"count" form:"count" json:"count" xml:"count" yaml:"count" binding:"required"`
	}
	e = h.Bind(c, &params)
	if e != nil {
		return
	}
	if params.Count < 1 {
		h.NegotiateErrorString(c, http.StatusBadRequest, `not support count`)
		return
	}
	dir, name := filepath.Split(filename)
	dir = filepath.Clean(dir + `/.chunks_` + name)
	f, e := os.Create(filename)
	if e != nil {
		h.NegotiateError(c, http.StatusInternalServerError, e)
		return
	}
	var s *os.File
	hash := md5.New()
	chunk := md5.New()
	buffer := make([]byte, 32)
	for i := 0; i < params.Count; i++ {
		s, e = os.Open(dir + `/` + fmt.Sprint(i))
		if e != nil {
			break
		}
		chunk.Reset()
		_, e = io.Copy(io.MultiWriter(chunk, f), s)
		s.Close()
		if e != nil {
			break
		}
		val := chunk.Sum(nil)
		hex.Encode(buffer, val)
		_, e = hash.Write(buffer)
		if e != nil {
			break
		}
	}
	f.Close()
	if e != nil {
		os.Remove(filename)
		h.NegotiateError(c, http.StatusInternalServerError, e)
		return
	}
	str := hex.EncodeToString(hash.Sum(nil))
	if str != params.Hash {
		os.Remove(filename)
		h.NegotiateErrorString(c, http.StatusInternalServerError, `hash not match`)
		return
	}

	os.RemoveAll(dir)
	c.Status(http.StatusCreated)
}
