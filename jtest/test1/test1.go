package test1

import "github.com/chroblert/jgoutils/jlog"

func init() {
	jlog.Info("init info: test1")
}

func Test1() {
	jlog.Debug("debug: test1")
	jlog.Info("info: test11")
}
