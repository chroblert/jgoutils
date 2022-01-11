package main

import (
	"github.com/chroblert/jgoutils/jasync"
	"github.com/chroblert/jgoutils/jlog"
	"strconv"
)

func main() {
	a := jasync.New()
	a.Add("t1", func() {
		jlog.NDebug("1")
	}, func() {
	})
	for i := 1; i < 100; i++ {
		t := i
		a.Add("t"+strconv.Itoa(i), func() int {
			return t
		}, func(s int) {
			jlog.NDebug(s)
		})
	}
	a.Run(-1)
	a.Wait()
}
