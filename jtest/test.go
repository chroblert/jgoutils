package main

import (
	"fmt"
	"github.com/chroblert/jgoutils/jlog"
	"github.com/chroblert/jgoutils/jtest/test1"
	"github.com/chroblert/jgoutils/jtest/test2"
	"os"
)

//func init(){
//	log.Println("test")
//}

func main() {
	jlog.SetLevel(jlog.DEBUG)
	jlog.Warn("warn: main")
	test1.Test1()
	test2.Test2()
	jlog.Println("xxx")
	jlog.Printf("%s\n", "testlll")
	fmt.Fprintln(os.Stderr, "xxxxxxx")
	jlog.NDebug("ndebug")
}
