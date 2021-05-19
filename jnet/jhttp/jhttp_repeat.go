package jhttp

import (
	"fmt"
	"github.com/chroblert/jgoutils/jasync"
	"github.com/chroblert/jgoutils/jlog"
	"github.com/chroblert/jgoutils/jrequests"
	"strconv"
)

func (hm *httpMsg) Repeat(counts ...int) (statuscode int, headers map[string][]string, body []byte, err error) {
	if len(counts) > 1 || (len(counts) == 1 && counts[0] < 1) {
		jlog.Fatal("请留空或只输入一位代表重放次数的正整数值")
	}
	if !hm.isUseSSL {
		hm.reqUrl = "http://" + hm.reqHost + hm.reqPath
	} else {
		hm.reqUrl = "https://" + hm.reqHost + hm.reqPath
	}
	if len(counts) == 0 || counts[0] == 1 {
		if hm.reqMethod == "GET" {
			return jrequests.Get(hm.reqUrl, jrequests.SetHeaders(hm.reqHeaders), jrequests.SetIsVerifySSL(hm.isVerifySSL), jrequests.SetParams(hm.reqParams), jrequests.SetData(hm.reqData), jrequests.SetProxy(hm.getProxy()))
		} else if hm.reqMethod == "POST" {
			return jrequests.Post(hm.reqUrl, jrequests.SetHeaders(hm.reqHeaders), jrequests.SetIsVerifySSL(hm.isVerifySSL), jrequests.SetParams(hm.reqParams), jrequests.SetData(hm.reqData), jrequests.SetProxy(hm.getProxy()))
		} else {
			return 0, nil, nil, fmt.Errorf("only GET or POST")
		}
	} else {
		jasyncobj := jasync.New()
		if hm.reqMethod == "GET" {
			for i := 0; i < counts[0]; i++ {
				jasyncobj.Add(strconv.Itoa(i), jrequests.Get, nil, hm.reqUrl, jrequests.SetHeaders(hm.reqHeaders), jrequests.SetIsVerifySSL(hm.isVerifySSL), jrequests.SetParams(hm.reqParams), jrequests.SetData(hm.reqData), jrequests.SetProxy(hm.getProxy()))
			}
		} else if hm.reqMethod == "POST" {
			for i := 0; i < counts[0]; i++ {
				jasyncobj.Add(strconv.Itoa(i), jrequests.Get, nil, hm.reqUrl, jrequests.SetHeaders(hm.reqHeaders), jrequests.SetIsVerifySSL(hm.isVerifySSL), jrequests.SetParams(hm.reqParams), jrequests.SetData(hm.reqData), jrequests.SetProxy(hm.getProxy()))
			}
		} else {
			return 0, nil, nil, fmt.Errorf("only GET or POST")
		}
		jasyncobj.Run()
		jasyncobj.Wait()
		jasyncobj.Clean()
		return 0, nil, nil, nil
	}

}
