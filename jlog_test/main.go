package main

import (
	"github.com/chroblert/jgoutils/jfile"
	"github.com/chroblert/jgoutils/jlog"
	"os"
	"path/filepath"
	"sync"
)

var (
	nlog *jlog.FishLogger
)

func main() {
	nlog = jlog.New()
	//nlog.SetLogFullPath("logs\\nlog.log")
	nlog.SetStoreToFile(false)
	defer func() {
		nlog.Flush()
	}()
	//jfile.ProcessLine("E:\\test-2000-2100.log", func(s string) error {
	//	nlog.NInfo(s)
	//	return nil
	//},false)
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	nlog.Info(exPath)
	nlog.Info(jfile.GetWorkPath())
	//p,_ := jfile.GetWorkPath()
	nlog.Info(jfile.GetAbsPath("fdfasd\\fsadfa\\../ddd/log"))
	nlog.Info(jfile.PathExists("ct95C4.tmp"))
	wg := sync.WaitGroup{}
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(t int) {
			defer wg.Done()
			nlog.Warn(t)
		}(i)
	}
	wg.Wait()
}
