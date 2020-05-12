package cookie

import (
	"encoding/hex"
	"encoding/json"
	"io/ioutil"

	"github.com/gorilla/securecookie"
	"gitlab.com/king011/king-go/os/fileperm"
)

type _Key struct {
	Hash  string
	Block string
}

// Generate generate random key
func Generate() (hashKey []byte, blockKey []byte) {
	hashKey = securecookie.GenerateRandomKey(32)
	blockKey = securecookie.GenerateRandomKey(32)
	return
}

// Save save key to file
func Save(filename string, hashKey []byte, blockKey []byte) (e error) {
	b, e := json.MarshalIndent(_Key{
		Hash:  hex.EncodeToString(hashKey),
		Block: hex.EncodeToString(blockKey),
	}, "", "\t")
	if e != nil {
		return
	}
	e = ioutil.WriteFile(filename, b, fileperm.File)
	if e != nil {
		return
	}
	return
}
