package test2

import "github.com/chroblert/jgoutils/jlog"

func init() {
	jlog.Debug("init debug: test2")
}

func Test2() {
	jlog.Debug("debug: test22")
	jlog.Info("info: test2")
}
