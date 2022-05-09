package jlog

import (
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
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

// newLogger 实例化logger
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

// newLogger 实例化logger
// path 日志完整路径 eg:logs/app.log
func newLogger(logConf LogConfig) *FishLogger {
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
	fl.iniCreateNewLog = logConf.InitCreateNewLog
	fl.storeToFile = logConf.StoreToFile

	fl.pool = sync.Pool{
		New: func() interface{} {
			return new(buffer)
		},
	}
	// 220509: 设置不将日志保存到文件
	if !fl.storeToFile {
		return fl
	}
	//日志文件路径设置
	fl.logFileExt = filepath.Ext(fl.logFullPath)                       // .log
	fl.logFileName = strings.TrimSuffix(fl.logFullPath, fl.logFileExt) // logs/app
	if fl.logFileExt == "" {
		fl.logFileExt = ".log"
	}
	os.MkdirAll(filepath.Dir(fl.logFullPath), 0666)

	signalChannel := make(chan os.Signal, 1)
	go fl.daemon(signalChannel)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)
	return fl
}

// 新建一个jlog示例
// 若不传入LogConfig，则使用默认的只进行创建。每十秒将日志写入文件，不限制存储天数和文件个数，单日志文件大小500MB，在控制台显示，每次运行不新建日志文件
// 否则，根据传入的LogConfig创建jlog示例
func New(logConfs ...LogConfig) *FishLogger {
	if len(logConfs) == 1 {
		logConf := logConfs[0]
		return newLogger(logConf)
	}
	return newLogger(LogConfig{
		BufferSize:        2048,
		FlushInterval:     10 * time.Second,
		MaxStoreDays:      -1,
		MaxSizePerLogFile: 512000000,
		LogCount:          -1,
		LogFullPath:       "logs/app.log",
		Lv:                DEBUG,
		UseConsole:        true,
		Verbose:           true,
		InitCreateNewLog:  false,
		StoreToFile:       true,
	})
}
