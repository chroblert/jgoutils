package jlog

import (
	"github.com/chroblert/jgoutils/jconfig"
	"time"
)

//
//var applog = new(FishLogger)
//
//// 20201011: 使用Logs
//func InitLogs(logpath string, amaxSize int64, amaxAge, alogCount int) {
//	maxSize = amaxSize // 单个文件最大大小
//	maxAge = amaxAge   // 单个文件保存2天
//	logCount = alogCount
//	applog = NewLogger(logpath)
//	defer applog.flush()
//	applog.setLevel(DEBUG)
//	applog.setVerbose(true)
//	applog.setConsole(true)
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
	//log.Println("jlog init")
}
