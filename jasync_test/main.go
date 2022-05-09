package main

import (
	"github.com/chroblert/jgoutils/jasync"
	"github.com/chroblert/jgoutils/jlog"
	"strconv"
	"time"
)

func main() {
	jlog.SetVerbose(false)
	a := jasync.New(true)
	a.Add("t1", func() (string, int) {
		jlog.NDebug("1xx")
		return "test 2", 22
	}, func() {
	})
	for i := 1; i < 5; i++ {
		t := i
		jlog.Debug(a.Add("t"+strconv.Itoa(i), func() []int {
			//jlog.Debug("t"+strconv.Itoa(t))
			return []int{1, t}
		}, func(s []int) {
			jlog.Debug("start ", s)
			time.Sleep(time.Duration(t) * time.Second)
			jlog.Debug(s)
		}))
	}
	a.Run(100)
	a.Wait()
	//a.Clean()
	a.Add("t12", func() (string, int) {
		jlog.NDebug("12xx")
		return "test 2", 22
	}, func(p1 string, p2 int) {
		time.Sleep(time.Duration(1) * time.Second)
		if p2 == 22 {
			jlog.Error("this is a test")
		}
	})
	a.Add("t13", func() (string, int) {
		jlog.NDebug("12xx")
		return "test 2", 22
	}, func(p1 string, p2 int) {
		time.Sleep(time.Duration(1) * time.Second)
		if p2 == 22 {
			jlog.Error("this is a test")
		}
	})
	jlog.Debug(a.GetTaskAllTotal())
	jlog.Debug(a.GetTaskCurAllTotal())
	jlog.Debug(a.Run(1))
	a.Wait()
	jlog.Debug(a.GetTasksResult())
	jlog.Debug(a.GetTaskAllTotal())
	jlog.Debug(a.GetTaskCurAllTotal())
	jlog.Debug(a.GetTaskCurNeedDoCount())
	a.PrintTaskStatus("t12", true)
	jlog.Debug("kkkkkkkjkj")
}
