package cookie

import (
	"encoding/hex"
	"io/ioutil"
	"os"
	"time"

	"gitlab.com/king011/webpc/helper"
	"gitlab.com/king011/webpc/logger"

	"github.com/gorilla/securecookie"
	"go.uber.org/zap"
)

var _Secure *securecookie.SecureCookie

// MaxAge return max age of timeout
func MaxAge() int64 {
	return _MaxAge
}

// Encode encode cookie
func Encode(name string, value interface{}) (string, error) {
	return _Secure.Encode(name, value)
}

// Decode decode  cookie
func Decode(name, value string, dst interface{}) error {
	return _Secure.Decode(name, value, dst)
}

// IsInit return is init ?
func IsInit() bool {
	return _Secure != nil
}

// IniClient client debug init
func IniClient(filename string) (e error) {
	b, e := ioutil.ReadFile(filename)
	if e != nil {
		return
	}

	var key _Key
	e = helper.Unmarshal(b, &key)
	if e != nil {
		return
	}
	hashKey, e := hex.DecodeString(key.Hash)
	if e != nil {
		return
	}
	blockKey, e := hex.DecodeString(key.Block)
	if e != nil {
		return
	}

	_Secure = securecookie.New(hashKey, blockKey)
	return
}

var _MaxAge int64

// Init initialize cookie system
func Init(filename string, maxAge time.Duration) (e error) {
	b, e := ioutil.ReadFile(filename)
	if e != nil {
		if os.IsNotExist(e) {
			e = nil
			newGenerate(filename, maxAge)
			return
		}
		if ce := logger.Logger.Check(zap.FatalLevel, "load securecookie"); ce != nil {
			ce.Write(
				zap.Error(e),
			)
		}
		return
	}
	var key _Key
	e = helper.Unmarshal(b, &key)
	if e != nil {
		if ce := logger.Logger.Check(zap.FatalLevel, "unmarshal securecookie"); ce != nil {
			ce.Write(
				zap.Error(e),
			)
		}
		return
	}
	hashKey, e := hex.DecodeString(key.Hash)
	if e != nil {
		return
	} else if len(hashKey) != 32 {
		if ce := logger.Logger.Check(zap.FatalLevel, "bad hash"); ce != nil {
			ce.Write(
				zap.String("key", key.Hash),
			)
		}
		return
	}
	blockKey, e := hex.DecodeString(key.Block)
	if e != nil {
		return
	} else if len(blockKey) != 32 {
		if ce := logger.Logger.Check(zap.FatalLevel, "bad block"); ce != nil {
			ce.Write(
				zap.String("key", key.Block),
			)
		}
		return
	}
	initKey(hashKey, blockKey, maxAge)
	return
}
func initKey(hashKey, blockKey []byte, maxAge time.Duration) {
	_Secure = securecookie.New(hashKey, blockKey)
	_MaxAge = int64(maxAge / time.Second)
	_Secure.MaxAge((int)(_MaxAge))
	if ce := logger.Logger.Check(zap.InfoLevel, "cookie"); ce != nil {
		ce.Write(
			zap.String("timeout", maxAge.String()),
		)
	}
	return
}
func newGenerate(filename string, maxAge time.Duration) {
	hashKey, blockKey := Generate()
	e := Save(filename, hashKey, blockKey)
	if e != nil {
		if ce := logger.Logger.Check(zap.FatalLevel, "save securecookie"); ce != nil {
			ce.Write(
				zap.Error(e),
			)
		}
	}
	initKey(hashKey, blockKey, maxAge)
	return
}
