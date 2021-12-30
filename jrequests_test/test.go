package main

import (
	"github.com/chroblert/jgoutils/jlog"
	"github.com/chroblert/jgoutils/jrequests"
	"strconv"
	"sync"
)

func main() {
	req, _ := jrequests.New()
	//req.SetIsVerifySSL(false)
	req.SetProxy("http://localhost:8080")
	req.SetHttpVersion(2)
	req.SetKeepalive(true)
	resp, err := req.Get("https://ipinfo.io")
	if err != nil {
		jlog.Error(err)
		return
	}
	jlog.Info(resp.Resp.Header)
	jlog.Info(resp.Resp.ProtoMajor)
	//jlog.Info(string(resp.Body()))
	var wg = &sync.WaitGroup{}
	for i := 0; i < 1; i++ {
		wg.Add(1)
		go func(t int) {
			defer wg.Done()
			resp, err := jrequests.CGet("http://myip.ipip.net/").CSetParams(map[string][]string{strconv.Itoa(t): {strconv.Itoa(t)}}).CSetData(nil).CSetProxy("http://localhost:8080").CSetParams(map[string][]string{"q1": {"v1 '\"", "v2"}}).CSetHeaders(map[string][]string{"Content-Type": {"application/json"}, "Accept": {"application/json"}}).CDo()
			if err != nil {
				jlog.Error(err)
				return
			}
			jlog.Info(string(resp.Body()))
		}(i)
	}
	wg.Wait()
	resp, err = jrequests.CGet("http://myip.ipip.net").CAddParams(map[string]string{"qqq1": "fdsa"}).CSetData(nil).CSetProxy("http://localhost:8080").CSetParams(map[string][]string{"q1": {"v1 '\"", "v2"}}).CSetHeaders(map[string][]string{"Content-Type": {"application/json"}, "Acceptxxx": {"application/json"}}).CAddHeaders(map[string]string{"kkkk": "kdfjadlksjf"}).CAddParams(map[string]string{"fadsfas": "{fasdfsad"}).CDo()
	if err != nil {
		jlog.Error(err)
		return
	}
	jlog.Info(string(resp.Body()))
}
