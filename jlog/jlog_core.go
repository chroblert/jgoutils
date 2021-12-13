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
//func NewLogger_old(fullPath string) *FishLogger {
//	fl := new(FishLogger)
//	fl.LogFullPath = fullPath                                    // logs/app.log
//	fl.logFileExt = filepath.Ext(fullPath)                       // .log
//	fl.logFileName = strings.TrimSuffix(fullPath, fl.logFileExt) // logs/app
//	if fl.logFileExt == "" {
//		fl.logFileExt = ".log"
//	}
//	os.MkdirAll(filepath.Dir(fullPath), 0666)
//	fl.level = DEBUG
//	fl.MaxStoreDays = maxAge
//	fl.MaxSizePerLogFile = maxSize
//	fl.pool = sync.Pool{
//		New: func() interface{} {
//			return new(buffer)
//		},
//	}
//	siganlChannel := make(chan os.Signal, 1)
//	go fl.daemon(siganlChannel)
//	signal.Notify(siganlChannel, syscall.SIGINT, syscall.SIGTERM)
//	return fl
//}

// NewLogger 实例化logger
// path 日志完整路径 eg:logs/app.log
func NewLogger(logConf LogConfig) *FishLogger {
	fl := new(FishLogger)
	// 日志配置
	fl.bufferSize = logConf.BufferSize
	fl.flushInterval = logConf.FlushInterval
	fl.maxStoreDays = logConf.MaxStoreDays
	fl.maxSizePerLogFile = logConf.MaxSizePerLogFile
	fl.logCount = logConf.LogCount
	fl.logFullPath = logConf.LogFullPath // logs/app.log
	fl.level = logConf.Lv
	fl.console = logConf.UseConsole
	fl.verbose = logConf.Verbose
	//日志文件路径设置
	fl.logFileExt = filepath.Ext(fl.logFullPath)                       // .log
	fl.logFileName = strings.TrimSuffix(fl.logFullPath, fl.logFileExt) // logs/app
	if fl.logFileExt == "" {
		fl.logFileExt = ".log"
	}
	os.MkdirAll(filepath.Dir(fl.logFullPath), 0666)
	fl.pool = sync.Pool{
		New: func() interface{} {
			return new(buffer)
		},
	}
	signalChannel := make(chan os.Signal, 1)
	go fl.daemon(signalChannel)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)
	return fl
}
