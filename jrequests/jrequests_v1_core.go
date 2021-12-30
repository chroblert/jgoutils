// +build ignore

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
	"time"
)

/*
发起单个请求

:param method: 请求方式, GET,POST,...
:param reqUrl: 请求url
:param cookies: cookie
:param headers: 请求头
:param proxy: 代理
:param data: Post请求body体
:param params: Get请求query string
:param isredirect: 是否跳转
:param timeout: 超时设置
:return statuscode: 状态码
:return headers: 响应头
:return body: 响应体
*/
//func SingleReq(method, reqUrl string,cookies map[string]string,headers map[string]string,proxy string,data []byte,params map[string]string, isredirect bool,timeout int)(statuscode int,respheaders map[string][]string,body []byte,err error){
func SingleReq(method, reqUrl string, option Option) (statuscode int, respheaders map[string][]string, body []byte, err error) {
	var client *http.Client
	client = &http.Client{}
	var httpTransport *http.Transport
	httpTransport = &http.Transport{}
	//var http2Transport *http2.transport
	//http2Transport = &http2.transport{}
	// 设置代理
	if option.Proxy != "" {
		proxy2 := func(_ *http.Request) (*url.URL, error) {
			return url.Parse(option.Proxy)
		}
		httpTransport.Proxy = proxy2
	}
	// 设置是否验证服务端证书
	if !option.IsVerifySSL {
		httpTransport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true, // 遇到不安全的https跳过验证
		}
	} else {
		var rootCAPool *x509.CertPool
		if rootCAPool, err = x509.SystemCertPool(); err != nil {
			rootCAPool = x509.NewCertPool()
		}
		// 判断当前程序运行的目录下是否有cas目录
		// 根证书，用来验证服务端证书的ca
		if isExsit, _ := jfile.PathExists(option.CAPath); isExsit {
			// 枚举当前目录下的文件
			caFilenames, _ := jfile.GetFilenamesByDir(option.CAPath)
			if len(caFilenames) > 0 {
				for _, filename := range caFilenames {
					caCrt, err := ioutil.ReadFile(filename)
					if err != nil {
						return -1, nil, nil, err
					}
					//jlog.Debug("导入证书结果:", rootCAPool.AppendCertsFromPEM(caCrt))
					rootCAPool.AppendCertsFromPEM(caCrt)
				}
			}
		}
		httpTransport.TLSClientConfig = &tls.Config{
			RootCAs: rootCAPool,
		}
	}
	// 设置httptransport
	switch option.HttpVersion {
	case 1:
		client.Transport = httpTransport
	case 2:
		// 升级到http2
		http2.ConfigureTransport(httpTransport)
		client.Transport = httpTransport
	}
	// 设置超时
	if option.Timeout != 0 {
		client.Timeout = time.Second * time.Duration(option.Timeout)
	}
	// 设置是否转发
	if !option.IsRedirect {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
	var req *http.Request
	var reader io.Reader
	if option.Data != nil {
		reader = bytes.NewReader(option.Data)
	} else {
		reader = nil
	}
	req, err = http.NewRequest(method, reqUrl, reader)
	if err != nil {
		//jlog.Error("http.NewRequest,error: ", err)
		return -1, nil, nil, err
	}
	// setting request params
	if option.Params != nil {
		q := req.URL.Query()
		for paramKey, paramValue := range option.Params {
			q.Add(paramKey, paramValue)
		}
		req.URL.RawQuery = q.Encode()
	}
	// 设置header
	if option.Headers != nil {
		for headerKey, headerValue := range option.Headers {
			req.Header.Add(headerKey, headerValue)
		}
	}
	// 设置cookiejar
	client.Jar, _ = cookiejar.New(nil)
	// 设置cookie
	var cookieList []*http.Cookie
	if option.Cookies != nil {
		for cookieKey, cookieValue := range option.Cookies {
			cookieList = append(cookieList, &http.Cookie{Name: cookieKey, Value: cookieValue})
		}
	}
	u, err := url.Parse(reqUrl)
	if err != nil {
		return -1, nil, nil, err
	}
	client.Jar.SetCookies(u, cookieList)
	// 设置是否使用长连接，是否为keep-alive,默认false
	req.Close = !option.IsKeepAlive
	// 发送请求
	resp, err := client.Do(req)
	//return resp, err
	if err != nil {
		//jlog.Error("client.Do,error: ", err)
		return -1, nil, nil, err

	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return -1, nil, nil, err

	}
	return resp.StatusCode, resp.Header, body, nil
}
