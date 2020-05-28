package mount

import (
	"gitlab.com/king011/webpc/configure"
	"gitlab.com/king011/webpc/logger"
	"go.uber.org/zap"
)

const (
	// CmdError 錯誤
	CmdError = iota + 1
	// CmdHeart websocket 心跳防止瀏覽器 關閉不獲取 websocket
	CmdHeart
	// CmdProgress 更新進度
	CmdProgress
	// CmdDone 操作完成
	CmdDone
	// CmdInit 初始化
	CmdInit
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
