package jhttp_test

import (
	"github.com/chroblert/jgoutils/jasync"
	"github.com/chroblert/jgoutils/jnet/jhttp"
	"github.com/chroblert/jgoutils/jrequests"
	"github.com/hashicorp/go-uuid"
	"testing"
)

func TestRepeat(t *testing.T) {
	j := jhttp.New()
	j.SetProxy("http://localhost:8080/kkkkkkkkk")
	j.InitWithFile("F:\\test2.txt")
	j.Repeat(1)

}

func BenchmarkRepeat(b *testing.B) {
	a := jasync.New()
	for i := 0; i < b.N; i++ {
		tt, _ := uuid.GenerateUUID()
		a.Add("", jrequests.Get, nil, "http://localhost:8000/?"+tt, jrequests.SetProxy("http://localhost:8080"))
	}
	a.Run(-1)
	a.Wait()
	//a.Run(-1)
}
