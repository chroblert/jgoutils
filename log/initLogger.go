package log

import "os"

// "github.com/chroblert/JCRandomProxy/v3/Conf"

var applog = new(FishLogger)

// @title InitLogs
// @description: 初始化logger
// @auth JC0o0l
// @param logpath string "日志路径"
// @param amaxSize int64 "单个文件最大大小,单位B"
// @param amaxAge int "单个文件保存天数"
// @param alogCount int "日志文件最大个数"
// @return 无
func InitLogs(logpath string, amaxSize int64, amaxAge, alogCount int) {
	maxSize = amaxSize // 单个文件最大大小
	maxAge = amaxAge   // 单个文件保存2天
	logCount = alogCount
	applog = NewLogger(logpath)
	defer applog.Flush()
	applog.SetLevel(DEBUG)
	applog.SetCallInfo(true)
	applog.SetConsole(true)
}

/*
Println in log
*/
func Println(args ...interface{}) {
	// applog.Info(args)
	applog.println(INFO, args...)
}

/*
Printf in log
*/
func Printf(format string, args ...interface{}) {
	// applog.Infof(format, args...)
	applog.printf(INFO, format, args...)
}

func Debug(args ...interface{}) {
	applog.println(DEBUG, args...)
}

func Debugf(format string, args ...interface{}) {
	applog.printf(DEBUG, format, args...)
}
func Info(args ...interface{}) {
	applog.println(INFO, args...)
}

func Infof(format string, args ...interface{}) {
	applog.printf(INFO, format, args...)
}

func Warn(args ...interface{}) {
	applog.println(WARN, args...)
}

func Warnf(format string, args ...interface{}) {
	applog.printf(WARN, format, args...)
}

func Error(args ...interface{}) {
	applog.println(ERROR, args...)
}

func Errorf(format string, args ...interface{}) {
	applog.printf(ERROR, format, args...)
}

func Fatal(args ...interface{}) {
	applog.println(FATAL, args...)
	os.Exit(0)
}
func Fatalf(format string, args ...interface{}) {
	applog.printf(FATAL, format, args...)
	os.Exit(0)
}
