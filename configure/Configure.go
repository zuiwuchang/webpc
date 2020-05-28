package configure

import (
	"io/ioutil"

	"github.com/google/go-jsonnet"
	logger "gitlab.com/king011/king-go/log/logger.zap"
	"gitlab.com/king011/webpc/helper"
)

// Configure global configure
type Configure struct {
	HTTP     HTTP
	System   System
	Cookie   Cookie
	Logger   logger.Options
	basePath string
	filename string
}

// Format format global configure
func (c *Configure) Format() (e error) {
	if e = c.HTTP.Format(c.basePath); e != nil {
		return
	}
	if e = c.System.Format(c.basePath); e != nil {
		return
	}
	if e = c.Cookie.Format(c.basePath); e != nil {
		return
	}
	return
}
func (c *Configure) String() string {
	if c == nil {
		return "nil"
	}
	b, e := helper.MarshalIndent(c, "", "	")
	if e != nil {
		return e.Error()
	}
	return string(b)
}

var _Configure Configure

// Single single Configure
func Single() *Configure {
	return &_Configure
}

// BasePath .
func (c *Configure) BasePath() string {
	return c.basePath
}

// Load load configure file
func (c *Configure) Load(basePath, filename string) (e error) {
	var b []byte
	b, e = ioutil.ReadFile(filename)
	if e != nil {
		return
	}
	vm := jsonnet.MakeVM()
	var jsonStr string
	jsonStr, e = vm.EvaluateSnippet("", string(b))
	if e != nil {
		return
	}
	b = []byte(jsonStr)
	e = helper.Unmarshal(b, c)
	if e != nil {
		return
	}
	c.basePath = basePath
	c.filename = filename
	return
}
