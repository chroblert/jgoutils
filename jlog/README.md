## install
go get -u github.com/chroblert/jgoutils/jlog
## Use
### 直接使用
打印日志有如下方法:
jlog.Debug()
jlog.Debugf()
jlog.Info()
jlog.Infof()
jlog.Warn()
jlog.Warnf()
jlog.Error()
jlog.Errorf()
jlog.Fatal()
jlog.Fatalf()
设置日志:
jlog.SetLogFullPath("logs/app.log"): 设置文件路径
jlog.SetLogCount(5): 设置日志文件保存的数量
jlog.SetLogLevel(0): 设置日志等级(低于该等级的日志不输出)
jlog.SetUseConsole(true): 设置是否在控制台打印日志
jlog2.Flush(): 写入文件 // 主程序结束前调用
### 新建实例使用
jlog2 := jlog.NewLogger(jlog.LogConfig{
BufferSize:        2048,
FlushInterval:     10*time.Second,
MaxStoreDays:      5,
MaxSizePerLogFile: 204800000,
LogCount:          5,
LogFullPath:       "logs/app2.log",
Lv: jlog.DEBUG,
UseConsole: true,
})
