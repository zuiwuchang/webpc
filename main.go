package main

import (
	"log"

	_ "gitlab.com/king011/webpc/assets/en-US/statik"
	_ "gitlab.com/king011/webpc/assets/zh-Hans/statik"
	_ "gitlab.com/king011/webpc/assets/zh-Hant/statik"
	"gitlab.com/king011/webpc/cmd"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if e := cmd.Execute(); e != nil {
		log.Fatalln(e)
	}
}
