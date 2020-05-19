package mount

import (
	"gitlab.com/king011/webpc/configure"
	"gitlab.com/king011/webpc/logger"
	"go.uber.org/zap"
)

var fs FileSystem

// Init .
func Init(ms []configure.Mount) (e error) {
	count := len(ms)
	for i := 0; i < count; i++ {
		fs.Push(ms[i].Name, ms[i].Root, ms[i].Read, ms[i].Write, ms[i].Shared)
		if ce := logger.Logger.Check(zap.InfoLevel, `mount`); ce != nil {
			ce.Write(
				zap.String(`name`, ms[i].Name),
				zap.String(`root`, ms[i].Root),
				zap.Bool(`read`, ms[i].Read),
				zap.Bool(`write`, ms[i].Write),
				zap.Bool(`shared`, ms[i].Shared),
			)
		}
	}
	return
}
