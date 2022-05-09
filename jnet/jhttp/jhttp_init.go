package jhttp

import "github.com/chroblert/jgoutils/jlog"

var jHttpLog = jlog.New(jlog.LogConfig{
	BufferSize:        0,
	FlushInterval:     0,
	MaxStoreDays:      0,
	MaxSizePerLogFile: 0,
	LogCount:          0,
	LogFullPath:       "",
	Lv:                0,
	UseConsole:        false,
	Verbose:           false,
	InitCreateNewLog:  false,
	StoreToFile:       false,
})

func init() {
	jHttpLog.SetStoreToFile(false)
}
