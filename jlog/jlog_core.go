package jlog

import (
	"github.com/chroblert/JC-GoUtils/jconfig"
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

	digits   = "0123456789"
	logShort = "[D][I][W][E][F]"
)

var (
	bufferSize          = jconfig.Conf.LogConfig.BufferSize // 256 KB
	flushInterval       = time.Duration(jconfig.Conf.LogConfig.FlushInterval) * time.Second
	maxAge              = jconfig.Conf.LogConfig.MaxStoreDays // 180 天
	maxSize       int64 = jconfig.Conf.LogConfig.MaxSize      // 256 MB
	logCount            = jconfig.Conf.LogConfig.LogCount
	fishLogger          = NewLogger(jconfig.Conf.LogConfig.LogFileName) // 默认实例
)

func init() {
	SetVerbose(true)
	SetLevel(logLevel(jconfig.Conf.LogConfig.LV))
	SetConsole(jconfig.Conf.LogConfig.IsConsole)
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
