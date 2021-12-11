package jlog

import (
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
)

func init() {
	//log.Println("jlog core")
}

// 字符串等级
func (lv logLevel) Str() string {
	if lv >= DEBUG && lv <= FATAL {
		return logShort[lv*3 : lv*3+3]
	}
	return "[N]"
}

// NewLogger 实例化logger
// path 日志完整路径 eg:logs/app.log
func NewLogger(fullPath string) *FishLogger {
	fl := new(FishLogger)
	fl.fullLogFilePath = fullPath                                // logs/app.log
	fl.logFileExt = filepath.Ext(fullPath)                       // .log
	fl.logFileName = strings.TrimSuffix(fullPath, fl.logFileExt) // logs/app
	if fl.logFileExt == "" {
		fl.logFileExt = ".log"
	}
	os.MkdirAll(filepath.Dir(fullPath), 0666)
	fl.level = DEBUG
	fl.maxStoreDays = maxAge
	fl.maxSizePerLogFile = maxSize
	fl.pool = sync.Pool{
		New: func() interface{} {
			return new(buffer)
		},
	}
	siganlChannel := make(chan os.Signal, 1)
	go fl.daemon(siganlChannel)
	signal.Notify(siganlChannel, syscall.SIGINT, syscall.SIGTERM)
	return fl
}
