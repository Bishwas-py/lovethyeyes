package main

import (
	"github.com/Bishwas-py/notify"
	"path"
	"runtime"
)

func main() {
	logo := "logo.png"
	_, filename, _, _ := runtime.Caller(0)
	logoDir := path.Join(path.Dir(filename), logo)
	notify.Notify("app name", "notice", "some text", logoDir)
}
