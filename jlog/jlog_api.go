package jlog

import "os"

// 设置

func SetVerbose(b bool) {
	fishLogger.setVerbose(b)
}

// 设置控制台输出
func SetConsole(b bool) {
	fishLogger.setConsole(b)
}

// 设置实例等级
func SetLevel(lv logLevel) {
	fishLogger.setLevel(lv)
}

// 设置最大保存天数
// 小于0不删除
func SetMaxStoreDays(ma int) {
	fishLogger.setMaxStoreDays(ma)
}

// -------- 实例 fishLogger
func Println(args ...interface{}) {
	fishLogger.println(11, args...)
}
func Printf(format string, args ...interface{}) {
	fishLogger.printf(11, format, args...)
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
	fishLogger.flush()
	os.Exit(0)
}
func Fatalf(format string, args ...interface{}) {
	fishLogger.printf(FATAL, format, args...)
	fishLogger.flush()
	os.Exit(0)
}


// 写入文件
func Flush() {
	fishLogger.flush()
}

