package jlog

import (
	"os"
)

func init() {
	//log.Println("jlog api")
}

// 设置

func SetVerbose(b bool) {
	fishLogger.setVerbose(b)
}

// 设置控制台输出
func SetUseConsole(b bool) {
	fishLogger.SetUseConsole(b)
}

// 设置实例等级
func SetLevel(lv logLevel) {
	fishLogger.SetLogLevel(lv)
}

// 设置最大保存天数
// 小于0不删除
func SetMaxStoreDays(ma int) {
	fishLogger.SetMaxStoreDays(ma)
}

// 设置日志文件保存数量
func SetLogCount(logCount int) {
	fishLogger.SetLogCount(logCount)
}

// 第一次运行是否新建日志
func IsIniCreateNewLog(iniCreateNewLog bool) {
	fishLogger.IsIniCreateNewLog(iniCreateNewLog)
}

// 设置文件保存路径
func SetLogFullPath(logFullPath string) {
	fishLogger.SetLogFullPath(logFullPath)
}

func SetMaxSizePerLogFile(logfilesize int64) {
	fishLogger.SetMaxSizePerLogFile(logfilesize)
}

// 设置是否保存到文件
func SetStoreToFile(b bool) {
	fishLogger.SetStoreToFile(b)
}

// -------- 实例 fishLogger
func Println(args ...interface{}) {
	fishLogger.nprintln(DEBUG, args...)
}
func Printf(format string, args ...interface{}) {
	fishLogger.nprintf(DEBUG, format, args...)
}

func Debug(args ...interface{}) {
	fishLogger.println(DEBUG, args...)
}

func Debugf(format string, args ...interface{}) {
	fishLogger.printf(DEBUG, format, args...)
}
func Info(args ...interface{}) {
	fishLogger.println(INFO, args...)
}

func Infof(format string, args ...interface{}) {
	fishLogger.printf(INFO, format, args...)
}

func Warn(args ...interface{}) {
	fishLogger.println(WARN, args...)
}

func Warnf(format string, args ...interface{}) {
	fishLogger.printf(WARN, format, args...)
}

func Error(args ...interface{}) {
	fishLogger.println(ERROR, args...)
}

func Errorf(format string, args ...interface{}) {
	fishLogger.printf(ERROR, format, args...)
}

func Fatal(args ...interface{}) {
	fishLogger.println(FATAL, args...)
	fishLogger.Flush()
	os.Exit(0)
}
func Fatalf(format string, args ...interface{}) {
	fishLogger.printf(FATAL, format, args...)
	fishLogger.Flush()
	os.Exit(0)
}

// 写入文件
func Flush() {
	//fmt.Println("size1:",fishLogger.writer.Buffered())
	fishLogger.Flush()
	//fmt.Println("size2:",fishLogger.writer.Size())

}

func NDebug(args ...interface{}) {
	fishLogger.nprintln(DEBUG, args...)
}

func NDebugf(format string, args ...interface{}) {
	fishLogger.nprintf(DEBUG, format, args...)
}
func NInfo(args ...interface{}) {
	fishLogger.nprintln(INFO, args...)
}

func NInfof(format string, args ...interface{}) {
	fishLogger.nprintf(INFO, format, args...)
}

func NWarn(args ...interface{}) {
	fishLogger.nprintln(WARN, args...)
}

func NWarnf(format string, args ...interface{}) {
	fishLogger.nprintf(WARN, format, args...)
}

func NError(args ...interface{}) {
	fishLogger.nprintln(ERROR, args...)
}

func NErrorf(format string, args ...interface{}) {
	fishLogger.nprintf(ERROR, format, args...)
}

func NFatal(args ...interface{}) {
	fishLogger.nprintln(FATAL, args...)
	fishLogger.Flush()
	os.Exit(0)
}
func NFatalf(format string, args ...interface{}) {
	fishLogger.nprintf(FATAL, format, args...)
	fishLogger.Flush()
	os.Exit(0)
}
