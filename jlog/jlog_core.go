package jlog

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	DEBUG logLevel = iota
	INFO
	WARN
	ERROR
	FATAL

	bufferSize    = 1024 * 256 // 256 KB
	digits        = "0123456789"
	flushInterval = 5 * time.Second
	logShort      = "[D][I][W][E][F]"
)

var (
	maxAge         = 180               // 180 天
	maxSize  int64 = 1024 * 1024 * 256 // 256 MB
	logCount       = 5
	fishLogger = NewLogger("logs/app.log")	// 默认实例
)


func init(){
	SetVerbose(true)
	SetLevel(DEBUG)
	SetConsole(true)
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
	go fl.daemon()
	return fl
}





