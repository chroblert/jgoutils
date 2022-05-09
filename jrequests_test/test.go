package main

import (
	"fmt"
	"github.com/chroblert/jgoutils/jasync"
	"github.com/chroblert/jgoutils/jlog"
	"github.com/chroblert/jgoutils/jrequests"
)

func main() {
	req, _ := jrequests.New()
	//req.SetIsVerifySSL(false)
	req.SetProxy("http://localhost:8080")
	req.SetIsVerifySSL(false)
	req.SetHttpVersion(2)
	req.SetKeepalive(false)
	_, err := req.Get("https://myip.ipip.net")
	if err != nil {
		jlog.Error(err)
		return
	}
	a := jasync.New()
	a.Add("", jrequests.CGet("https://myip.ipip.net/11").CAddHeaders(map[string]string{"kkkk": "t====ddd"}).CSetIsVerifySSL(false).CSetHttpVersion(2).CSetProxy("http://localhost:8080").CSetTimeout(3).CDo, nil)
	a.Add("", jrequests.CGet("https://myip.ipip.net/12").CAddHeaders(map[string]string{"ddd": "t====ddd"}).CSetIsVerifySSL(false).CSetHttpVersion(2).CSetProxy("http://localhost:8080").CSetTimeout(3).CDo, nil)
	req.AddHeaders(map[string]string{"test": "header test"})
	for i := 3; i < 6; i++ {
		a.Add("", req.Get, nil, "https://myip.ipip.net?"+fmt.Sprintf("%d", i))
	}
	a.Run(-1)
	a.Wait()
	a.PrintAllTaskStatus(true)
	return
	req.SetHttpVersion(1)
	_, err = req.Get("https://ipinfo.io")
	if err != nil {
		jlog.Error(err)
		return
	}
	//req.SetHttpVersion(2)
	_, err = req.Get("https://myip.ipip.net")
	if err != nil {
		jlog.Error(err)
		return
	}
	_, err = jrequests.CGet("https://ipinfo.io").CSetIsVerifySSL(false).CSetHttpVersion(2).CSetProxy("http://localhost:8080").CSetTimeout(3).CDo()
	if err != nil {
		jlog.Error(err)
		//jlog.NFatal("not find arcsight ArcMC in",target)
		return
	}

}
