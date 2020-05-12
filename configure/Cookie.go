package configure

import (
	"gitlab.com/king011/webpc/utils"
	"time"
)

// Cookie configure cookie
type Cookie struct {
	Filename string
	MaxAge   time.Duration
}

// Format .
func (c *Cookie) Format(basePath string) (e error) {
	if c.Filename == "" {
		c.Filename = "securecookie.json"
	}
	c.Filename = utils.Abs(basePath, c.Filename)
	if c.MaxAge < time.Second {
		c.MaxAge = time.Hour * 24
	} else {
		c.MaxAge *= time.Millisecond
	}
	return
}
