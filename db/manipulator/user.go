package manipulator

import (
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"

	"gitlab.com/king011/webpc/cookie"
	"gitlab.com/king011/webpc/utils"

	"github.com/boltdb/bolt"
	"gitlab.com/king011/webpc/db/data"
	"gitlab.com/king011/webpc/logger"
	"go.uber.org/zap"
)

// User 用戶 操縱器
type User struct {
}

// Init 初始化 bucket
func (m User) Init(tx *bolt.Tx) (e error) {
	bucket, e := tx.CreateBucketIfNotExists([]byte(data.UserBucket))
	if e != nil {
		return
	}
	cursor := bucket.Cursor()
	k, _ := cursor.First()
	if k != nil {
		return
	}
	if ce := logger.Logger.Check(zap.WarnLevel, `init default user`); ce != nil {
		ce.Write(
			zap.String(`name`, `killer`),
			zap.String(`password`, `19890604`),
		)
	}
	if !logger.Logger.OutConsole() {
		fmt.Println(`init default user name=killer password=19890604`)
	}
	password := sha512.Sum512([]byte("19890604"))
	u := data.User{
		Name:     `killer`,
		Password: hex.EncodeToString(password[:]),
		Shell:    true,
		Read:     true,
		Write:    true,
		Root:     true,
	}
	b, e := u.Encoder()
	if e != nil {
		return
	}
	e = bucket.Put([]byte("killer"), b)
	return
}

// Login 登入
func (m User) Login(name, password string) (result *cookie.Session, e error) {
	e = _db.View(func(t *bolt.Tx) (e error) {
		bucket := t.Bucket([]byte(data.UserBucket))
		if bucket == nil {
			e = fmt.Errorf("bucket not exist : %s", data.UserBucket)
			return
		}
		val := bucket.Get(utils.StringToBytes(name))
		if val == nil {
			return
		}
		var u data.User
		e = u.Decode(val)
		if e != nil {
			return
		}

		if u.Password == password {
			result = &cookie.Session{
				Name:  name,
				Shell: u.Shell,
				Read:  u.Read,
				Write: u.Write,
				Root:  u.Root,
			}
		}
		return
	})
	return
}

// List 返回用戶 列表
func (m User) List() (result []*data.User, e error) {
	e = _db.View(func(t *bolt.Tx) (e error) {
		bucket := t.Bucket([]byte(data.UserBucket))
		if bucket == nil {
			e = fmt.Errorf("bucket not exist : %s", data.UserBucket)
			return
		}
		bucket.ForEach(func(k, v []byte) error {
			var u data.User
			e = u.Decode(v)
			if e == nil {
				result = append(result, &u)
			}
			return nil
		})
		return
	})
	return
}

// Add 添加用戶
func (m User) Add(name, password string, shell, read, write, root bool) (e error) {
	if name == "" {
		e = errors.New("name not support empty")
		return
	}
	if password == "" {
		e = errors.New("password not support empty")
		return
	}
	e = _db.Update(func(t *bolt.Tx) (e error) {
		bucket := t.Bucket([]byte(data.UserBucket))
		if bucket == nil {
			e = fmt.Errorf("bucket not exist : %s", data.UserBucket)
			return
		}
		key := utils.StringToBytes(name)
		v := bucket.Get(key)
		if v != nil {
			e = fmt.Errorf("user already exists : %s", name)
			return
		}
		u := data.User{
			Name:     name,
			Password: password,
			Shell:    shell,
			Read:     read,
			Write:    write,
			Root:     root,
		}
		b, e := u.Encoder()
		if e != nil {
			return
		}
		e = bucket.Put(key, b)
		if e != nil {
			return
		}
		// 新建 shell
		bucket = t.Bucket([]byte(data.ShellBucket))
		if bucket == nil {
			e = fmt.Errorf("bucket not exist : %s", data.ShellBucket)
			return
		}
		_, e = bucket.CreateBucket(key)
		return
	})
	return
}

// Remove 刪除用戶
func (m User) Remove(name string) (e error) {
	if name == "" {
		e = errors.New("name not support empty")
		return
	}
	e = _db.Update(func(t *bolt.Tx) (e error) {
		// 删除用户
		bucket := t.Bucket([]byte(data.UserBucket))
		if bucket == nil {
			e = fmt.Errorf("bucket not exist : %s", data.UserBucket)
			return
		}
		key := utils.StringToBytes(name)
		e = bucket.Delete(key)
		if e != nil {
			return
		}

		// 删除 shell
		bucket = t.Bucket([]byte(data.ShellBucket))
		if bucket == nil {
			e = fmt.Errorf("bucket not exist : %s", data.ShellBucket)
			return
		}
		e = bucket.DeleteBucket(key)
		if e == bolt.ErrBucketNotFound {
			e = nil
			return
		}
		return
	})
	return
}

// Password3 修改密碼
func (m User) Password3(name, old, val string) (e error) {
	if name == "" {
		e = errors.New("name not support empty")
		return
	}
	if old == "" {
		e = errors.New("old password not support empty")
		return
	}
	if val == "" {
		e = errors.New("new password not support empty")
		return
	}
	e = _db.Update(func(t *bolt.Tx) (e error) {
		bucket := t.Bucket([]byte(data.UserBucket))
		if bucket == nil {
			e = fmt.Errorf("bucket not exist : %s", data.UserBucket)
			return
		}
		key := utils.StringToBytes(name)
		v := bucket.Get(key)
		if v == nil {
			e = fmt.Errorf("user not exists : %s", name)
			return
		}
		var u data.User
		e = u.Decode(v)
		if e != nil {
			return
		}
		if u.Password != old {
			e = errors.New("old password not match")
			return
		}
		if u.Password == val {
			return
		}
		u.Password = val
		b, e := u.Encoder()
		if e != nil {
			return
		}
		e = bucket.Put(key, b)
		return
	})
	return
}

// Password 修改密碼
func (m User) Password(name, password string) (e error) {
	if name == "" {
		e = errors.New("name not support empty")
		return
	}
	if password == "" {
		e = errors.New("password not support empty")
		return
	}
	e = _db.Update(func(t *bolt.Tx) (e error) {
		bucket := t.Bucket([]byte(data.UserBucket))
		if bucket == nil {
			e = fmt.Errorf("bucket not exist : %s", data.UserBucket)
			return
		}
		key := utils.StringToBytes(name)
		v := bucket.Get(key)
		if v == nil {
			e = fmt.Errorf("user not exists : %s", name)
			return
		}
		var u data.User
		e = u.Decode(v)
		if e != nil {
			return
		}
		if u.Password == password {
			return
		}
		u.Password = password
		b, e := u.Encoder()
		if e != nil {
			return
		}
		e = bucket.Put(key, b)
		return
	})
	return
}

// Change 更改權限
func (m User) Change(name string, shell, read, write, root bool) (e error) {
	if name == "" {
		e = errors.New("name not support empty")
		return
	}
	e = _db.Update(func(t *bolt.Tx) (e error) {
		bucket := t.Bucket([]byte(data.UserBucket))
		if bucket == nil {
			e = fmt.Errorf("bucket not exist : %s", data.UserBucket)
			return
		}
		key := utils.StringToBytes(name)
		v := bucket.Get(key)
		if v == nil {
			e = fmt.Errorf("user not exists : %s", name)
			return
		}
		var u data.User
		e = u.Decode(v)
		if e != nil {
			return
		}
		if u.Shell == shell &&
			u.Read == read && u.Write == write &&
			u.Root == root {
			return
		}
		u.Shell = shell
		u.Read = read
		u.Write = write
		u.Root = root
		b, e := u.Encoder()
		if e != nil {
			return
		}
		e = bucket.Put(key, b)
		return
	})
	return
}
