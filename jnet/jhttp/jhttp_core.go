package jhttp

import (
	"fmt"
	"github.com/chroblert/JC-GoUtils/jconfig"
	"github.com/chroblert/JC-GoUtils/jrequests"
	"strings"
)

type httpMsg struct {
	reqMethod   string
	reqHost     string
	reqUrl      string
	reqPath     string
	reqParams   map[string]string
	reqHeaders  map[string]string
	reqData     []byte
	isVerifySSL bool
	isUseSSL    bool
}

func New() *httpMsg {
	return &httpMsg{
		reqMethod:   "Get",
		reqHost:     "",
		reqUrl:      "/",
		reqPath:     "/",
		reqParams:   make(map[string]string),
		reqHeaders:  make(map[string]string),
		reqData:     make([]byte, 0),
		isVerifySSL: false,
	}
}

func (hm *httpMsg) InitWithFile(filename string) {
	hm.parseFromBurpReqFile(filename)
}

func (hm *httpMsg) InitWithBytes(reqMsg []byte) {
	hm.parseFromBytes(reqMsg)
}

func (hm *httpMsg) getInfoFromReqLine(reqLine []string) (reqMethod, reqPath string, reqParams map[string]string) {
	reqMethod = reqLine[0]
	reqParams = make(map[string]string)
	if strings.Index(reqLine[1], "?") != -1 {
		reqPath = reqLine[1][:strings.Index(reqLine[1], "?")]
		queryString := reqLine[1][strings.Index(reqLine[1], "?")+1:]
		for _, param := range strings.Split(queryString, "&") {
			idx := strings.Index(param, "=")
			reqParams[param[:idx]] = param[idx+1:]
		}
	} else {
		reqPath = reqLine[1]

	}
	return reqMethod, reqPath, reqParams
}

// 设置目标，如：http://test.test
func (hm *httpMsg) SetHost(target string) {
	if strings.Contains(target, "https") {
		hm.isVerifySSL = true
	}
	hm.reqHost = target[strings.Index(target, "/")+2:]

}

func (hm *httpMsg) SetIsVerifySSL(b bool) {
	hm.isVerifySSL = b
}
func (hm *httpMsg) SetIsUseSSL(b bool) {
	hm.isUseSSL = b
}
func (hm *httpMsg) Repeat() (statuscode int, headers map[string][]string, body []byte, err error) {
	if !hm.isUseSSL {
		hm.reqUrl = "http://" + hm.reqHost + hm.reqPath
	} else {
		hm.reqUrl = "https://" + hm.reqHost + hm.reqPath
	}
	if hm.reqMethod == "GET" {
		return jrequests.Get(hm.reqUrl, jrequests.SetHeaders(hm.reqHeaders), jrequests.SetIsVerifySSL(hm.isVerifySSL), jrequests.SetParams(hm.reqParams), jrequests.SetData(hm.reqData), jrequests.SetProxy(jconfig.Conf.RequestsConfig.Proxy))
	} else if hm.reqMethod == "POST" {
		return jrequests.Post(hm.reqUrl, jrequests.SetHeaders(hm.reqHeaders), jrequests.SetIsVerifySSL(hm.isVerifySSL), jrequests.SetParams(hm.reqParams), jrequests.SetData(hm.reqData), jrequests.SetProxy(jconfig.Conf.RequestsConfig.Proxy))
	}
	return 0, nil, nil, fmt.Errorf("only GET or POST")
}
