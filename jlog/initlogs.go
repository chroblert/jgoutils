package jlog

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
