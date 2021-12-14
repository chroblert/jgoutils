package jrequests

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"github.com/chroblert/jgoutils/jfile"
	"golang.org/x/net/http2"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"
)

var jrePool *sync.Pool = &sync.Pool{New: func() interface{} {
	return &jrequest{
		Proxy:   "",
		Timeout: 60,
		Headers: map[string][]string{
			"User-Agent": {"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:86.0) Gecko/20100101 Firefox/86.0"},
		},
		Data:        nil,
		Params:      nil,
		Cookies:     nil,
		IsRedirect:  false,
		IsVerifySSL: false,
		HttpVersion: 1,
		IsKeepAlive: false,
		CAPath:      "cas",
		//Url:         "",
		transport: &http.Transport{},
		cli:       &http.Client{},
	}
}}

// header,cookie,params
type jrequest struct {
	Headers map[string][]string
	Params  map[string][]string
	Cookies []*http.Cookie

	Proxy       string //func(*http.Request) (*url.URL, error)
	Timeout     int
	Data        []byte
	IsRedirect  bool
	IsVerifySSL bool
	HttpVersion int
	IsKeepAlive bool
	CAPath      string
	Url         string
	transport   *http.Transport
	cli         *http.Client
	req         *http.Request
	method      string
}

type requestConfig struct {
	Proxy       string //func(*http.Request) (*url.URL, error)
	Timeout     int
	Data        []byte
	IsRedirect  bool
	IsVerifySSL bool
	HttpVersion int
	IsKeepAlive bool
	CAPath      string
}
type jresponse struct {
	Resp *http.Response
}

// 返回响应的body
func (jrs *jresponse) Body() []byte {
	defer jrs.Resp.Body.Close()
	res, err := ioutil.ReadAll(jrs.Resp.Body)
	if err != nil {
		return nil
	}
	return res
}

// 创建实例
func New() (jr *jrequest, err error) {
	//jr = &jrequest{
	//	Proxy:       "",
	//	Timeout:     60,
	//	Headers:     map[string]string{
	//		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:86.0) Gecko/20100101 Firefox/86.0",
	//	},
	//	Data:        nil,
	//	Params:      nil,
	//	Cookies:     nil,
	//	IsRedirect:  false,
	//	IsVerifySSL: false,
	//	HttpVersion: 1,
	//	IsKeepAlive: false,
	//	CAPath:      "cas",
	//	//Url:         "",
	//	transport:   &http.Transport{},
	//	cli:         &http.Client{},
	//}
	jr = jrePool.Get().(*jrequest)
	jr.cli.Jar, err = cookiejar.New(nil)
	if err != nil {
		return
	}
	return
}

func (jr *jrequest) Do(d ...interface{}) (resp *jresponse, err error) {
	resp = &jresponse{}
	//jlog.Info(jr.req)
	resp.Resp, err = jr.cli.Do(jr.req)
	return
}

//func (jr *jrequest) A_Get(reqUrl string,d ...interface{}) (jre *jrequest){
//	//resp = &jresponse{}
//	//jr.Url = reqUrl
//	var reader io.Reader
//	if len(d) > 0{
//		switch d[0].(type) {
//		case []byte:
//			reader = bytes.NewReader(d[0].([]byte))
//		case string:
//			reader = strings.NewReader(d[0].(string))
//		default:
//			reader = nil
//		}
//	}else{
//		reader = nil
//	}
//	var err error
//	jr.req,err = http.NewRequest("GET",reqUrl,reader)
//	if err != nil{
//		return nil
//	}
//	// 设置headers
//	for k,v := range jr.Headers{
//		jr.req.Header.Add(k,v)
//	}
//	// 设置cookies
//	u,err := url.Parse(reqUrl)
//	jr.cli.Jar.SetCookies(u,jr.Cookies)
//	// 设置params
//	if jr.Params != nil {
//		query := jr.req.URL.Query()
//		for paramKey, paramValue := range jr.Params {
//			//query.Add(paramKey, paramValue)
//			for _,v2 := range paramValue{
//				query.Add(paramKey,v2)
//			}
//		}
//		jr.req.URL.RawQuery = query.Encode()
//	}
//	// 设置transport
//	jr.cli.Transport = jr.transport
//	// 设置connection
//	jr.req.Close = !jr.IsKeepAlive
//	//resp.Resp,err = jr.cli.Do(jr.req)
//	return jr
//}

func (jr *jrequest) Get(reqUrl string, d ...interface{}) (resp *jresponse, err error) {
	resp = &jresponse{}
	//jr.Url = reqUrl
	var reader io.Reader
	if len(d) > 0 {
		switch d[0].(type) {
		case []byte:
			reader = bytes.NewReader(d[0].([]byte))
		case string:
			reader = strings.NewReader(d[0].(string))
		default:
			reader = nil
		}
	} else {
		reader = nil
	}

	jr.req, err = http.NewRequest("GET", reqUrl, reader)
	if err != nil {
		return nil, err
	}
	// 设置headers
	for k, v := range jr.Headers {
		for _, v2 := range v {
			jr.req.Header.Add(k, v2)
		}
	}
	// 设置cookies
	u, err := url.Parse(reqUrl)
	jr.cli.Jar.SetCookies(u, jr.Cookies)
	// 设置params
	if jr.Params != nil {
		query := jr.req.URL.Query()
		for paramKey, paramValue := range jr.Params {
			//query.Add(paramKey, paramValue)
			for _, v2 := range paramValue {
				query.Add(paramKey, v2)
			}
		}
		jr.req.URL.RawQuery = query.Encode()
	}
	// 设置transport
	jr.cli.Transport = jr.transport
	// 设置connection
	jr.req.Close = !jr.IsKeepAlive
	resp.Resp, err = jr.cli.Do(jr.req)
	jr = &jrequest{
		Proxy:   "",
		Timeout: 60,
		Headers: map[string][]string{
			"User-Agent": {"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:86.0) Gecko/20100101 Firefox/86.0"},
		},
		Data:        nil,
		Params:      nil,
		Cookies:     nil,
		IsRedirect:  false,
		IsVerifySSL: false,
		HttpVersion: 1,
		IsKeepAlive: false,
		CAPath:      "cas",
		//Url:         "",
		transport: &http.Transport{},
		cli:       &http.Client{},
	}
	jrePool.Put(jr)
	return
}

func (jr *jrequest) Post(reqUrl string, d ...interface{}) (resp *jresponse, err error) {
	resp = &jresponse{}
	//jr.Url = reqUrl
	var reader io.Reader
	if len(d) > 0 {
		switch d[0].(type) {
		case []byte:
			reader = bytes.NewReader(d[0].([]byte))
		case string:
			reader = strings.NewReader(d[0].(string))
		default:
			reader = nil
		}
	} else {
		reader = nil
	}

	jr.req, err = http.NewRequest("POST", reqUrl, reader)
	if err != nil {
		return nil, err
	}
	// 设置headers
	for k, v := range jr.Headers {
		for _, v2 := range v {
			jr.req.Header.Add(k, v2)
		}
	}
	// 设置cookies
	u, err := url.Parse(reqUrl)
	jr.cli.Jar.SetCookies(u, jr.Cookies)
	// 设置params
	if jr.Params != nil {
		query := jr.req.URL.Query()
		for paramKey, paramValue := range jr.Params {
			//query.Add(paramKey, paramValue)
			for _, v2 := range paramValue {
				query.Add(paramKey, v2)
			}
		}
		jr.req.URL.RawQuery = query.Encode()
	}
	// 设置transport
	jr.cli.Transport = jr.transport
	// 设置connection
	jr.req.Close = !jr.IsKeepAlive
	resp.Resp, err = jr.cli.Do(jr.req)
	return
}

// 设置代理
func (jr *jrequest) SetProxy(proxy string) {
	// TODO proxy格式校验
	_, err := url.Parse(proxy)
	if err != nil {
		//jr.transport.Proxy = nil
		//jlog.Error(err)
		return
	}
	jr.Proxy = proxy
	if proxy != "" {
		jr.transport.Proxy = func(request *http.Request) (*url.URL, error) {
			return url.Parse(proxy)
		}
	} else {
		jr.transport.Proxy = nil
	}

}

// 设置超时
func (jr *jrequest) SetTimeout(timeout int) {
	jr.Timeout = timeout
	jr.cli.Timeout = time.Second * time.Duration(jr.Timeout)
}

// 重置并设置headers
func (jr *jrequest) SetHeaders(headers map[string][]string) {
	if jr == nil {
		return
	}
	if len(headers) == 0 {
		jr.Headers = make(map[string][]string)
		return
	} else {
		jr.Headers = make(map[string][]string, len(headers))
	}
	for k, v := range headers {
		jr.Headers[k] = make([]string, len(v))
		for k2, v2 := range v {
			jr.Headers[k][k2] = v2
		}
	}
}

// 添加headers
func (jr *jrequest) AddHeaders(headers map[string]string) {
	if jr == nil {
		return
	}
	if jr.Headers == nil {
		if len(headers) == 0 {
			jr.Headers = make(map[string][]string)
			return
		} else {
			jr.Headers = make(map[string][]string, len(headers))
		}
	}
	for k, v := range headers {
		if _, ok := jr.Headers[k]; !ok {
			jr.Headers[k] = []string{v}
		} else {
			jr.Headers[k] = append(jr.Headers[k], v)
		}
	}
}

// 设置body data
func (jr *jrequest) SetData(data []byte) {
	jr.Data = data
}

// 设置params
func (jr *jrequest) SetParams(params map[string][]string) {
	if jr.Params == nil {
		if len(params) == 0 {
			jr.Params = make(map[string][]string)
			return
		} else {
			jr.Params = make(map[string][]string, len(params))
		}

	}
	for k, v := range params {
		jr.Params[k] = make([]string, len(v))
		for k2, v2 := range v {
			jr.Params[k][k2] = v2
		}
	}
}

// 追加params,1
func (jr *jrequest) AddParams(params map[string]string) {
	if jr == nil {
		return
	}
	if jr.Params == nil {
		if len(params) == 0 {
			jr.Params = make(map[string][]string)
			return
		} else {
			jr.Params = make(map[string][]string, len(params))
		}
	}
	//jr.Params = params
	for k, v := range params {
		if _, ok := jr.Params[k]; !ok {
			jr.Params[k] = []string{v}
		} else {
			jr.Params[k] = append(jr.Params[k], v)
		}
	}
	return
}

// 设置cookies
func (jr *jrequest) SetCookies(cookies []map[string]string) {
	if jr.Cookies == nil {
		jr.Cookies = make([]*http.Cookie, len(cookies))
	}
	for k, cookie := range cookies {
		for k2, v2 := range cookie {
			jr.Cookies[k] = &http.Cookie{Name: k2, Value: v2}
			break
		}
	}
}

// 设置是否转发
func (jr *jrequest) SetIsRedirect(isredirect bool) {
	jr.IsRedirect = isredirect
	// 设置是否转发
	if !jr.IsRedirect {
		jr.cli.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
}

// 设置http 2.0
func (jr *jrequest) SetHttpVersion(version int) {
	jr.HttpVersion = version
	// 设置httptransport
	switch jr.HttpVersion {
	case 1:
		//client.transport = httpTransport
	case 2:
		// 升级到http2
		http2.ConfigureTransport(jr.transport)
		//client.transport = httpTransport
	}
}

// 设置是否验证ssl
func (jr *jrequest) SetIsVerifySSL(isverifyssl bool) {
	jr.IsVerifySSL = isverifyssl
	// 设置是否验证服务端证书
	if !jr.IsVerifySSL {
		jr.transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true, // 遇到不安全的https跳过验证
		}
	} else {
		var rootCAPool *x509.CertPool
		rootCAPool, err := x509.SystemCertPool()
		if err != nil {
			rootCAPool = x509.NewCertPool()
		}
		// 判断当前程序运行的目录下是否有cas目录
		// 根证书，用来验证服务端证书的ca
		if isExsit, _ := jfile.PathExists(jr.CAPath); isExsit {
			// 枚举当前目录下的文件
			caFilenames, _ := jfile.GetFilenamesByDir(jr.CAPath)
			if len(caFilenames) > 0 {
				for _, filename := range caFilenames {
					caCrt, err := ioutil.ReadFile(filename)
					if err != nil {
						return
					}
					//jlog.Debug("导入证书结果:", rootCAPool.AppendCertsFromPEM(caCrt))
					rootCAPool.AppendCertsFromPEM(caCrt)
				}
			}
		}
		jr.transport.TLSClientConfig = &tls.Config{
			RootCAs: rootCAPool,
		}
	}
}

// 设置connection是否为长连接，keep-alive
func (jr *jrequest) SetKeepalive(iskeepalive bool) {
	jr.IsKeepAlive = iskeepalive
}

// 设置capath
func (jr *jrequest) SetCAPath(CAPath string) {
	jr.CAPath = CAPath
}
