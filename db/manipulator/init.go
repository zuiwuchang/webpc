package manipulator

import (
	"os"

	"github.com/boltdb/bolt"
	"gitlab.com/king011/webpc/logger"
	"go.uber.org/zap"
)

type manipulator interface {
	Init(tx *bolt.Tx) (e error)
}

var _db *bolt.DB

// Init 初始化 數據庫
func Init(source string) (e error) {
	db, e := bolt.Open(source, 0600, nil)
	if e != nil {
		if ce := logger.Logger.Check(zap.FatalLevel, "open databases error"); ce != nil {
			ce.Write(
				zap.Error(e),
				zap.String("source", source),
			)
		}
		return
	}
	if ce := logger.Logger.Check(zap.InfoLevel, "open databases"); ce != nil {
		ce.Write(
			zap.String("source", source),
		)
	}

	e = db.Update(func(tx *bolt.Tx) (e error) {
		buckets := []manipulator{
			User{},
			Shell{},
		}
		for i := 0; i < len(buckets); i++ {
			e = buckets[i].Init(tx)
			if e != nil {
				return
			}
		}
		return
	})
	if e != nil {
		os.Exit(1)
	}
	_db = db
	return
}

// DB .
func DB() *bolt.DB {
	return _db
}
