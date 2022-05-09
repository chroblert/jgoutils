package jlog

import (
	"time"
)

//
//var applog = new(FishLogger)
//
//// 20201011: 使用Logs
//func InitLogs(logpath string, amaxSize int64, amaxAge, alogCount int) {
//	maxSize = amaxSize // 单个文件最大大小
//	maxAge = amaxAge   // 单个文件保存2天
//	LogCount = alogCount
//	applog = newLogger(logpath)
//	defer applog.Flush()
//	applog.SetLogLevel(DEBUG)
//	applog.setVerbose(true)
//	applog.SetUseConsole(true)
//	//applog.info("test")
//}
//func Println(args ...interface{}) {
//	// applog.info(args)
//	applog.println(INFO, args...)
//}
//
//func Printf(format string, args ...interface{}) {
//	// applog.infof(format, args...)
//	applog.printf(INFO, format, args...)
//}

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
	//BufferSize          = jconfig.Conf.LogConfig.BufferSize // 256 KB
	//FlushInterval       = time.Duration(jconfig.Conf.LogConfig.FlushInterval) * time.Second
	//maxAge              = jconfig.Conf.LogConfig.MaxStoreDays // 180 天
	//maxSize       int64 = jconfig.Conf.LogConfig.MaxSize      // 256 MB
	//LogCount            = jconfig.Conf.LogConfig.LogCount
	//fishLogger           = newLogger(jconfig.Conf.LogConfig.LogFileName) // 默认实例
	fishLogger = newLogger(LogConfig{
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
	})
)

//var jlogConfig = map[string]interface{}{
//	"BufferSize":2048,
//	"FlushInterval":10*time.Second,
//	"MaxStoreDays":5,
//	"MaxSizePerLogFile":20480000,
//	"LogCount":5,
//
//}

type LogConfig struct {
	BufferSize        int
	FlushInterval     time.Duration
	MaxStoreDays      int
	MaxSizePerLogFile int64
	LogCount          int
	LogFullPath       string
	Lv                logLevel
	UseConsole        bool
	Verbose           bool
	InitCreateNewLog  bool
	StoreToFile       bool
}

func init() {
	//SetVerbose(true)
	//SetLevel(logLevel(jconfig.Conf.LogConfig.LV))
	//SetConsole(jconfig.Conf.LogConfig.IsConsole)
	//log.Println("jlog init")
}
