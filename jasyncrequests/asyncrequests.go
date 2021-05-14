package jasyncrequests

import (
	"fmt"
	"github.com/chroblert/JC-GoUtils/jrequests"
	"time"
)

const HttpProxy  = "http://127.0.0.1:8080"

type Asyncresponse struct{
	Key string
	TimeElapsed float64
	Statuscode int
	Headers map[string][]string
	Body []byte
	Err string
}

func Fetch(keystr,requrl string, ch chan<- Asyncresponse) {
	asyncResp := Asyncresponse{Key:keystr}
	start := time.Now()
	statuscode,headers,body, err := jrequests.Get(requrl, jrequests.SetProxy(HttpProxy), jrequests.SetTimeout(5), jrequests.SetData([]byte("你好")))
	secs := time.Since(start).Seconds()
	asyncResp.TimeElapsed = secs
	if err != nil {
		asyncResp.Err = fmt.Sprint(err)
		ch <- asyncResp// send to channel ch
		return
	}
	asyncResp.Statuscode = statuscode
	asyncResp.Headers = headers
	asyncResp.Body = body
	ch <- asyncResp
}