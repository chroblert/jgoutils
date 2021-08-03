package jhttp

import (
	"github.com/chroblert/jgoutils/jconfig"
	"github.com/chroblert/jgoutils/jlog"
	"net/url"
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

	intruData *intruderData
	proxy     string
	//reqBytes []byte // 请求报文的字节数组
	//wordFiles []string // 字典文件切片
}
type intruderData struct {
	reqBytes  []byte   // 请求报文的字节数组
	wordFiles []string // 字典文件切片
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

		intruData: &intruderData{
			reqBytes:  make([]byte, 0),
			wordFiles: make([]string, 0),
		},
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

// 设置请求方法
func (hm *httpMsg) SetReqMethod(reqMethod string) {
	hm.reqMethod = reqMethod
}

// 设置请求体
func (hm *httpMsg) SetReqData(reqDataStr string) {
	hm.reqData = []byte(reqDataStr)
}

// 设置单个header
func (hm *httpMsg) SetHeader(header map[string]string) {
	for k, v := range header {
		if hm.reqHeaders[k] == "" {
			hm.reqHeaders[k] = v
		} else {
			hm.reqHeaders[k] = hm.reqHeaders[k] + "; " + v
		}
	}
}

// 设置URL，包含querystring
func (hm *httpMsg) SetURL(requrl string) error{
	urlobj, err := url.ParseRequestURI(requrl)
	if err != nil {
		jlog.Error("错误", err)
		return err
	}
	//jlog.Debug(urlobj.Scheme)
	//jlog.Debug(urlobj.Host)
	//jlog.Debug(urlobj.Path)
	//jlog.Debug(urlobj.Query())
	//jlog.Debug(urlobj.ForceQuery)
	if urlobj.Path == "" {
		hm.reqPath = "/"
	} else {
		hm.reqPath = urlobj.Path
	}
	for k, v := range urlobj.Query() {
		hm.reqParams[k] = strings.Join(v, "")
	}
	hm.reqHost = urlobj.Host
	if strings.Contains(urlobj.Scheme, "https") {
		hm.isUseSSL = true
	} else {
		hm.isUseSSL = false
	}
	return nil
	//jlog.Debug(hm.reqParams)
}

// 设置是否验证SSL
func (hm *httpMsg) SetIsVerifySSL(b bool) {
	hm.isVerifySSL = b
}

// 设置目标站点是否使用SSL
func (hm *httpMsg) SetIsUseSSL(b bool) {
	hm.isUseSSL = b
}

// 设置目标站点使用的代理
func (hm *httpMsg) SetProxy(proxy string) {
	hm.proxy = proxy
}

// 获取代理
func (hm *httpMsg) getProxy() string {
	if hm.proxy != "" {
		return hm.proxy
	} else if jconfig.Conf.RequestsConfig.Proxy != "" {
		return jconfig.Conf.RequestsConfig.Proxy
	} else {
		return ""
	}
}

// 设置暴破用的字典所在的文件
func (hm *httpMsg) SetWordfiles(wordfiles ...string) {
	for _, v := range wordfiles {
		hm.intruData.wordFiles = append(hm.intruData.wordFiles, v)
	}
	//jlog.Debug(hm.intruData.wordFiles)
}

func (hm *httpMsg) Clean() {
	hm.reqMethod = "Get"
	hm.reqHost = ""
	hm.reqUrl = "/"
	hm.reqPath = "/"
	hm.reqParams = make(map[string]string)
	hm.reqHeaders = make(map[string]string)
	hm.reqData = make([]byte, 0)
	hm.isVerifySSL = false

	hm.intruData.reqBytes = make([]byte, 0)
	hm.intruData.wordFiles = make([]string, 0)
}
