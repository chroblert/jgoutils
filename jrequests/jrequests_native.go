package jrequests

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"github.com/chroblert/JC-GoUtils/jconfig"
	"github.com/chroblert/JC-GoUtils/jfile"
	"github.com/chroblert/JC-GoUtils/jlog"
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
	// 设置代理
	var httpTransport *http.Transport
	httpTransport = &http.Transport{}
	if option.Proxy != "" {
		proxy2 := func(_ *http.Request) (*url.URL, error) {
			return url.Parse(option.Proxy)
		}
		httpTransport.Proxy = proxy2
	}
	// 设置是否验证ssl
	if !option.IsVerifySSL {
		//jlog.Info("jinlaiyaxxxxxx")
		httpTransport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true, // 遇到不安全的https跳过验证
		}
	} else {
		// 判断当前程序运行的目录下是否有cas目录
		jlog.Debug(jconfig.Conf.RequestsConfig.CAPath)
		if isExsit, _ := jfile.PathExists(jconfig.Conf.RequestsConfig.CAPath); isExsit {
			// 枚举当前目录下的文件
			filenams, _ := jfile.GetFilenamesByDir(jconfig.Conf.RequestsConfig.CAPath)
			if len(filenams) > 0 {
				var clientCrtPool *x509.CertPool
				if clientCrtPool, err = x509.SystemCertPool(); err != nil {
					jlog.Error(err)
					clientCrtPool = x509.NewCertPool()
				}
				for _, filename := range filenams {
					jlog.Debug(filename)
					jlog.Debug("导入ca证书:", filename)
					caCrt, _ := ioutil.ReadFile(filename)
					jlog.Debug("导入证书结果:", clientCrtPool.AppendCertsFromPEM(caCrt))
				}
				httpTransport.TLSClientConfig = &tls.Config{
					RootCAs: clientCrtPool,
				}
			}
		}
	}
	// 设置httptransport
	if httpTransport != nil {
		//jlog.Debug("使用httptransport:",httpTransport.TLSClientConfig)
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
	// setting request body
	var reader io.Reader
	if option.Data != nil {
		reader = bytes.NewReader(option.Data)
	} else {
		reader = nil
	}
	req, err = http.NewRequest(method, reqUrl, reader)
	if err != nil {
		jlog.Error("http.NewRequest,error: ", err)
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
	u, _ := url.Parse(reqUrl)
	client.Jar.SetCookies(u, cookieList)
	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		jlog.Error("client.Do,error: ", err)
		return -1, nil, nil, err

	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		jlog.Error("ioutil.ReadAll,error: ", err)
		return -1, nil, nil, err

	}
	//fmt.Printf(resp.Header)
	//fmt.Println(resp.StatusCode)
	//for k,v := range resp.Header{
	//	fmt.Println(k,":",v)
	//}
	return resp.StatusCode, resp.Header, body, nil
}
