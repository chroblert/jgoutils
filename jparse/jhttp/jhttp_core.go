package jhttp

import (
	"bufio"
	"bytes"
	"github.com/chroblert/JC-GoUtils/jlog"
	"github.com/chroblert/JC-GoUtils/jrequests"
	"io"
	"os"
	"strings"
)

type httpMsg struct{
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

func New() *httpMsg{
	return &httpMsg{
		reqMethod:   "Get",
		reqHost:     "",
		reqUrl:      "/",
		reqPath:     "/",
		reqParams:   make(map[string]string),
		reqHeaders:  make(map[string]string),
		reqData:     make([]byte,0),
		isVerifySSL: false,
	}
}

func (hm *httpMsg) Init(filename string){
	hm.parseFromBurpReqFile(filename)
}

func (hm *httpMsg) parseFromBurpReqFile(filename string)(reqLine []string,reqHeaders map[string]string,reqData []byte){
	f,_ := os.OpenFile(filename,os.O_RDONLY,0666)
	reader := bufio.NewReader(f)
	// 读取请求行
	//reqLine := make([]string,3)
	jlog.Debug("请求行:")
	if data,err := reader.ReadBytes('\n');err == nil{
		//jlog.Debug(string(data[:len(data)-2]))
		reqLine = strings.Split(string(data[:len(data)-2])," ")
		hm.reqMethod,hm.reqPath,hm.reqParams = hm.getInfoFromReqLine(reqLine)
		//tmpByte := bytes.Split(data,[]byte{' '})
		jlog.Debug(reqLine)
	}

	// 读取请求头
	reqHeaders = make(map[string]string)
	jlog.Debug("请求头:")
	for data,err := reader.ReadBytes('\n'); err == nil || err == io.EOF;data,err = reader.ReadBytes('\n') {
		if err == io.EOF{
			jlog.Fatal("报文格式错误",data)
			break
		}
		if len(data)==2{
			jlog.Debug("blank line")
			break
		}else{
			jlog.Debug(string(data[:len(data)-2]))
			// 保存请求头
			headerName := string(data[:bytes.IndexRune(data,':')])
			reqHeaders[headerName] = strings.TrimLeft(string(data[bytes.IndexRune(data,':')+1:len(data)-2])," ")
		}
	}
	jlog.Debug(reqHeaders)
	hm.reqHeaders = reqHeaders
	hm.reqHost = hm.reqHeaders["Host"]
	// 读取请求体
	//var reqData []byte
	jlog.Debug("请求体:")
	for data,err := reader.ReadBytes('\n'); err == nil || err == io.EOF;data,err = reader.ReadBytes('\n') {
		reqData = data
		if err == io.EOF{
			break
		}
		//jlog.Debug(string(data[:len(data)-2]))
	}
	jlog.Debug(reqData)
	hm.reqData = reqData
	return reqLine,reqHeaders,reqData

}

func (hm *httpMsg) getInfoFromReqLine(reqLine []string) (reqMethod,reqPath string, reqParams map[string]string){
	reqMethod = reqLine[0]
	reqParams = make(map[string]string)
	if strings.Index(reqLine[1],"?") != -1{
		reqPath = reqLine[1][:strings.Index(reqLine[1],"?")]
		queryString := reqLine[1][strings.Index(reqLine[1],"?")+1:]
		for _,param := range strings.Split(queryString,"&"){
			idx := strings.Index(param,"=")
			reqParams[param[:idx]]=param[idx+1:]
		}
	}else{
		reqPath = reqLine[1]

	}
	return reqMethod,reqPath,reqParams
}

// 设置目标，如：http://test.test
func (hm *httpMsg)SetHost(target string){
	if strings.Contains(target,"https"){
		hm.isVerifySSL = true
	}
	hm.reqHost = target[strings.Index(target,"/")+2:]

}

func (hm *httpMsg) SetIsVerifySSL(b bool){
	hm.isVerifySSL = b
}
func (hm *httpMsg) SetIsUseSSL(b bool){
	hm.isUseSSL = b
}
func (hm *httpMsg)Repeat(){
	if !hm.isUseSSL {
		hm.reqUrl = "http://" + hm.reqHost+hm.reqPath
	}else{
		hm.reqUrl = "https://" + hm.reqHost+hm.reqPath
	}
	if hm.reqMethod == "GET"{
		jrequests.Get(hm.reqUrl,jrequests.SetHeaders(hm.reqHeaders),jrequests.SetIsVerifySSL(hm.isVerifySSL),jrequests.SetParams(hm.reqParams),jrequests.SetData(hm.reqData))
	}else if hm.reqMethod == "POST"{
		jrequests.Post(hm.reqUrl,jrequests.SetHeaders(hm.reqHeaders),jrequests.SetIsVerifySSL(hm.isVerifySSL),jrequests.SetParams(hm.reqParams),jrequests.SetData(hm.reqData))
	}
}