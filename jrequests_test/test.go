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
	//var wg = &sync.WaitGroup{}
	//for i := 0; i< 1; i++{
	//	wg.Add(1)
	//	go func(t int) {
	//		defer wg.Done()
	//		resp,err := req.Get("https://ipinfo.io")
	//		if err != nil{
	//			jlog.Error(err)
	//			return
	//		}
	//		jlog.Error(t,resp)
	//		jlog.Info(t)
	//		jlog.Info(resp.Resp.Header)
	//		jlog.Info(resp.Resp.ProtoMajor)
	//		//jlog.Info(string(resp.Body()))
	//	}(i)
	//}
	//wg.Wait()
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
			resp, err := jrequests.A_Get("http://myip.ipip.net/").A_SetParams(map[string][]string{strconv.Itoa(t): {strconv.Itoa(t)}}).A_SetData(nil).A_SetProxy("http://localhost:8080").A_SetParams(map[string][]string{"q1": {"v1 '\"", "v2"}}).A_SetHeaders(map[string][]string{"Content-Type": {"application/json"}, "Accept": {"application/json"}}).A_Do()
			if err != nil {
				jlog.Error(err)
				return
			}
			jlog.Info(string(resp.Body()))
		}(i)
	}
	wg.Wait()
	resp, err = jrequests.A_Get("http://myip.ipip.net").A_AddParams(map[string]string{"qqq1": "fdsa"}).A_SetData(nil).A_SetProxy("http://localhost:8080").A_SetParams(map[string][]string{"q1": {"v1 '\"", "v2"}}).A_SetHeaders(map[string][]string{"Content-Type": {"application/json"}, "Acceptxxx": {"application/json"}}).A_AddHeaders(map[string]string{"kkkk": "kdfjadlksjf"}).A_AddParams(map[string]string{"fadsfas": "{fasdfsad"}).A_Do()
	if err != nil {
		jlog.Error(err)
		return
	}
	jlog.Info(string(resp.Body()))
}
