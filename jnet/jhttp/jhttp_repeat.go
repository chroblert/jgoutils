package jhttp

import (
	"fmt"
	"github.com/chroblert/jgoutils/jasync"
	"github.com/chroblert/jgoutils/jrequests"
)

func (hm *httpMsg) Repeat(counts ...int) map[string][]interface{} {
	if len(counts) > 1 || (len(counts) == 1 && counts[0] < 1) {
		jHttpLog.Error("请留空或只输入一位代表重放次数的正整数值")
		return nil
	}
	if !hm.isUseSSL {
		hm.reqUrl = "http://" + hm.reqHost + hm.reqPath
	} else {
		hm.reqUrl = "https://" + hm.reqHost + hm.reqPath
	}
	var asyncCount int
	if len(counts) == 0 || counts[0] == 1 {
		asyncCount = 1
	} else {
		asyncCount = counts[0]
	}

	jasyncobj := jasync.New()
	if hm.reqMethod == "GET" {
		for i := 0; i < asyncCount; i++ {
			jasyncobj.Add("", jrequests.Get, nil, hm.reqUrl, jrequests.SetHeaders(hm.reqHeaders), jrequests.SetIsVerifySSL(hm.isVerifySSL), jrequests.SetParams(hm.reqParams), jrequests.SetData(hm.reqData), jrequests.SetProxy(hm.getProxy()), jrequests.SetTimeout(hm.timeout))
		}
	} else if hm.reqMethod == "POST" {
		for i := 0; i < asyncCount; i++ {
			jasyncobj.Add("", jrequests.Post, nil, hm.reqUrl, jrequests.SetHeaders(hm.reqHeaders), jrequests.SetIsVerifySSL(hm.isVerifySSL), jrequests.SetParams(hm.reqParams), jrequests.SetData(hm.reqData), jrequests.SetProxy(hm.getProxy()), jrequests.SetTimeout(hm.timeout))
		}
	} else {
		return map[string][]interface{}{"0": []interface{}{0, nil, nil, fmt.Errorf("only GET or POST yet")}}
	}
	jasyncobj.Run(-1)
	jasyncobj.Wait()
	result := jasyncobj.GetTasksResult()
	jasyncobj.Clean()
	return result
	//}

}
