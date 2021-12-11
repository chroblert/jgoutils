package main

import (
	"fmt"
	"github.com/chroblert/jgoutils/jlog"
	"github.com/chroblert/jgoutils/jlog_test/test1"
	"github.com/chroblert/jgoutils/jlog_test/test2"
	"os"
	"sync"
	"time"
)

//func init(){
//	log.Println("test")
//}

func main() {
	jlog.SetLevel(jlog.DEBUG)
	//jlog.SetVerbose(false)
	jlog.Warn("warn: main")
	test1.Test1()
	test2.Test2()
	jlog.Println("xxx")
	jlog.Printf("%s\n", "testlll")
	fmt.Fprintln(os.Stderr, "xxxxxxx")
	jlog.NDebug("ndebug")
	var wg = &sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(t int) {
			jlog.Debug(t)
			wg.Done()
		}(i)
	}
	wg.Wait()

	jlog2 := jlog.NewLogger(jlog.LogConfig{
		BufferSize:        2048,
		FlushInterval:     10 * time.Second,
		MaxStoreDays:      5,
		MaxSizePerLogFile: 204800000,
		LogCount:          5,
		LogFullPath:       "logs/app2.log",
		Lv:                jlog.DEBUG,
		UseConsole:        true,
	})
	jlog2.Warn("jlog2 warn")
	jlog2.SetLogFullPath("logs/appdd.log")
	jlog2.NError("jlog2 error")
	//var wg2 = &sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(t int) {
			jlog2.Debug("jlog2", t)
			wg.Done()
		}(i)
	}
	wg.Wait()
	jlog.Flush()
	jlog2.Flush()

}
