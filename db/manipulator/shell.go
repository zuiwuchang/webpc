package manipulator

import (
	"fmt"

	"github.com/boltdb/bolt"
	"gitlab.com/king011/webpc/db/data"
	"gitlab.com/king011/webpc/utils"
)

// Shell shell 记录表
type Shell struct {
}

// Init 初始化 bucket
func (m Shell) Init(tx *bolt.Tx) (e error) {
	shell, e := tx.CreateBucketIfNotExists([]byte(data.ShellBucket))
	if e != nil {
		return
	}
	bucket := tx.Bucket(utils.StringToBytes(data.UserBucket))
	cursor := bucket.Cursor()
	for k, _ := cursor.First(); k != nil; k, _ = cursor.Next() {
		_, e = shell.CreateBucketIfNotExists(k)
		if e != nil {
			return
		}
	}
	return
}

// Add 添加记录
func (m Shell) Add(username string, element *data.Shell) (e error) {
	val, e := element.Encoder()
	if e != nil {
		return
	}
	e = _db.Update(func(t *bolt.Tx) (e error) {
		bucket := t.Bucket([]byte(data.ShellBucket))
		if bucket == nil {
			e = fmt.Errorf("bucket not exist : %s", data.ShellBucket)
			return
		}
		key := utils.StringToBytes(username)
		bucket = bucket.Bucket(key)
		if bucket == nil {
			e = fmt.Errorf("bucket not exist : %s.%s", data.ShellBucket, username)
			return
		}
		e = bucket.Put(data.EncoderID(uint64(element.ID)), val)
		return
	})
	return
}

// Remove 删除记录
func (m Shell) Remove(username string, id int64) (e error) {
	e = _db.Update(func(t *bolt.Tx) (e error) {
		bucket := t.Bucket([]byte(data.ShellBucket))
		if bucket == nil {
			e = fmt.Errorf("bucket not exist : %s", data.ShellBucket)
			return
		}
		key := utils.StringToBytes(username)
		bucket = bucket.Bucket(key)
		if bucket == nil {
			return
		}
		bucket.Delete(data.EncoderID(uint64(id)))
		return
	})
	return
}
