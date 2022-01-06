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

var jrePool = &sync.Pool{New: func() interface{} {
	return &jrequest{
		Proxy:   "",
		Timeout: 60,
		Headers: map[string][]string{
			"User-Agent": {"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:86.0) Gecko/20100101 Firefox/86.0"},
		},
		Data:         nil,
		Params:       nil,
		Cookies:      nil,
		IsRedirect:   true,
		IsVerifySSL:  false,
		HttpVersion:  1,
		IsKeepAlive:  false,
		IsKeepCookie: false,
		CAPath:       "cas",
		//Url:         "",
		transport:  &http.Transport{},
		transport2: &http2.Transport{},
		cli:        &http.Client{},
	}
}}

// 用于链式
type jrequest struct {
	Headers map[string][]string
	Params  map[string][]string
	Cookies []*http.Cookie

	Proxy        string //func(*http.Request) (*url.URL, error)
	Timeout      int
	Data         []byte
	IsRedirect   bool
	IsVerifySSL  bool
	HttpVersion  int
	IsKeepAlive  bool
	IsKeepCookie bool
	CAPath       string
	Url          string
	transport    *http.Transport
	transport2   *http2.Transport
	cli          *http.Client
	req          *http.Request
	method       string
}

// 用于新建
type jnrequest struct {
	Headers map[string][]string
	Params  map[string][]string
	Cookies []*http.Cookie

	Proxy        string //func(*http.Request) (*url.URL, error)
	Timeout      int
	Data         []byte
	IsRedirect   bool
	IsVerifySSL  bool
	HttpVersion  int
	IsKeepAlive  bool
	IsKeepCookie bool
	CAPath       string
	Url          string
	transport    *http.Transport
	transport2   *http2.Transport
	cli          *http.Client
	req          *http.Request
	method       string
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
// param d:是否保存cookie，true or false
func New(d ...interface{}) (jrn *jnrequest, err error) {
	jrn = (*jnrequest)(jrePool.Get().(*jrequest))
	jrn.cli.Jar, err = cookiejar.New(nil)
	if err != nil {
		return
	}
	// 设置是否保存cookie
	if len(d) > 0 {
		switch d[0].(type) {
		case bool:
			jrn.IsKeepCookie = d[0].(bool)
		default:
			jrn.IsKeepCookie = false
		}
	}
	return
}

//func (jr *jrequest) Do(d ...interface{}) (resp *jresponse, err error) {
//	resp = &jresponse{}
//	//jlog.Info(req2)
//	resp.Resp, err = jr.cli.Do(jr.req)
//	return
//}

// TODO 解决并发 资源共享问题
func (jr *jnrequest) Get(reqUrl string, d ...interface{}) (resp *jresponse, err error) {
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

	//req2, err = http.NewRequest("GET", reqUrl, reader)
	req2, err := http.NewRequest("GET", reqUrl, reader)
	if err != nil {
		return nil, err
	}
	// 设置headers
	for k, v := range jr.Headers {
		for _, v2 := range v {
			req2.Header.Add(k, v2)
		}
	}
	// 设置cookies
	u, err := url.Parse(reqUrl)
	jr.cli.Jar.SetCookies(u, jr.Cookies)
	// 设置是否转发
	if !jr.IsRedirect {
		jr.cli.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
	// 设置params
	if jr.Params != nil {
		query := req2.URL.Query()
		for paramKey, paramValue := range jr.Params {
			//query.Add(paramKey, paramValue)
			for _, v2 := range paramValue {
				query.Add(paramKey, v2)
			}
		}
		req2.URL.RawQuery = query.Encode()
	}
	// 设置transport
	// TODO 做个备份 没起作用??? new一次，只能为 http/1.1或http/2
	backTransport := jr.transport
	//tmp := *jr.transport
	//backTransport := &tmp
	if jr.HttpVersion == 2 {
		// 判断当前是否已经为http2
		alreadyH2 := false
		for _, v := range jr.transport.TLSClientConfig.NextProtos {
			if v == "h2" {
				alreadyH2 = true
				break
			}
		}
		if !alreadyH2 {
			err = http2.ConfigureTransport(backTransport)
			if err != nil {
				return nil, err
			}
		}
	}
	jr.cli.Transport = backTransport
	// 设置connection
	req2.Close = !jr.IsKeepAlive
	resp.Resp, err = jr.cli.Do(req2)
	// 清空cookie
	if err == nil {
		// 清空cookie
		if !jr.IsKeepCookie {
			jr.cli.Jar, err = cookiejar.New(nil)
		}
	}
	return
}

func resetJr(jr *jrequest) {
	jr.Proxy = ""
	jr.Timeout = 60
	jr.Headers = map[string][]string{
		"User-Agent": {"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:86.0) Gecko/20100101 Firefox/86.0"},
	}
	jr.Data = nil
	jr.Params = nil
	jr.Cookies = nil
	jr.IsRedirect = true
	jr.IsVerifySSL = false
	jr.HttpVersion = 1
	jr.IsKeepAlive = false
	jr.CAPath = "cas"
	jr.transport = &http.Transport{}
	jr.transport2 = &http2.Transport{}
	jr.cli = &http.Client{}
}

func (jr *jnrequest) Post(reqUrl string, d ...interface{}) (resp *jresponse, err error) {
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

	req2, err := http.NewRequest("POST", reqUrl, reader)
	if err != nil {
		return nil, err
	}
	// 设置headers
	for k, v := range jr.Headers {
		for _, v2 := range v {
			req2.Header.Add(k, v2)
		}
	}
	// 设置cookies
	u, err := url.Parse(reqUrl)
	jr.cli.Jar.SetCookies(u, jr.Cookies)
	// 设置是否转发
	if !jr.IsRedirect {
		jr.cli.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
	// 设置params
	if jr.Params != nil {
		query := req2.URL.Query()
		for paramKey, paramValue := range jr.Params {
			//query.Add(paramKey, paramValue)
			for _, v2 := range paramValue {
				query.Add(paramKey, v2)
			}
		}
		req2.URL.RawQuery = query.Encode()
	}
	// 设置transport
	// TODO 做个备份 没起作用??? new一次，只能为 http/1.1或http/2
	backTransport := jr.transport
	//tmp := *jr.transport
	//backTransport := &tmp
	if jr.HttpVersion == 2 {
		// 判断当前是否已经为http2
		alreadyH2 := false
		for _, v := range jr.transport.TLSClientConfig.NextProtos {
			if v == "h2" {
				alreadyH2 = true
				break
			}
		}
		if !alreadyH2 {
			err = http2.ConfigureTransport(backTransport)
			if err != nil {
				return nil, err
			}
		}
	}
	jr.cli.Transport = backTransport
	// 设置connection
	req2.Close = !jr.IsKeepAlive
	resp.Resp, err = jr.cli.Do(req2)
	if err == nil {
		// 清空cookie
		if !jr.IsKeepCookie {
			jr.cli.Jar, err = cookiejar.New(nil)
		}
	}
	return
}

// 设置代理
func (jr *jnrequest) SetProxy(proxy string) {
	if jr == nil {
		return
	}
	// TODO proxy格式校验
	_, err := url.Parse(proxy)
	if err != nil {
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
func (jr *jnrequest) SetTimeout(timeout int) {
	if jr == nil {
		return
	}
	jr.Timeout = timeout
	jr.cli.Timeout = time.Second * time.Duration(jr.Timeout)
}

// 重置并设置headers
func (jr *jnrequest) SetHeaders(headers map[string][]string) {
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
func (jr *jnrequest) AddHeaders(headers map[string]string) {
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
func (jr *jnrequest) SetData(d interface{}) {
	if jr == nil {
		return
	}
	switch d.(type) {
	case []byte:
		jr.Data = d.([]byte)
	case string:
		jr.Data = []byte(d.(string))
	default:
		jr.Data = []byte(nil)
	}
	//jr.Data = data
}

// 设置params
func (jr *jnrequest) SetParams(params map[string][]string) {
	if jr == nil {
		return
	}
	if len(params) == 0 {
		jr.Params = make(map[string][]string)
		return
	} else {
		jr.Params = make(map[string][]string, len(params))
	}
	for k, v := range params {
		jr.Params[k] = make([]string, len(v))
		for k2, v2 := range v {
			jr.Params[k][k2] = v2
		}
	}
}

// 追加params,1
func (jr *jnrequest) AddParams(params map[string]string) {
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
func (jr *jnrequest) SetCookies(cookies []map[string]string) {
	if jr == nil {
		return
	}
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
func (jr *jnrequest) SetIsRedirect(isredirect bool) {
	if jr == nil {
		return
	}
	jr.IsRedirect = isredirect
	// 设置是否转发
	if !jr.IsRedirect {
		jr.cli.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
}

// 设置http 2.0
func (jr *jnrequest) SetHttpVersion(version int) {
	if jr == nil {
		return
	}
	jr.HttpVersion = version
}

// 设置是否验证ssl
func (jr *jnrequest) SetIsVerifySSL(isverifyssl bool) {
	if jr == nil {
		return
	}
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
func (jr *jnrequest) SetKeepalive(iskeepalive bool) {
	if jr == nil {
		return
	}
	jr.IsKeepAlive = iskeepalive
}

// 设置capath
func (jr *jnrequest) SetCAPath(CAPath string) {
	if jr == nil {
		return
	}
	jr.CAPath = CAPath
}
