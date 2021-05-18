package jhttp

import (
	"bufio"
	"bytes"
	"github.com/chroblert/JC-GoUtils/jlog"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

func (hm *httpMsg) parseFromBurpReqFile(filename string) (reqLine []string, reqHeaders map[string]string, reqData []byte) {
	tmpBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		jlog.Fatal("读取文件错误")
	}
	// 添加至reqBytes
	hm.intruData.reqBytes = append(hm.intruData.reqBytes, tmpBytes...)

	// 下面解析http报文
	f, _ := os.OpenFile(filename, os.O_RDONLY, 0666)
	reader := bufio.NewReader(f)
	// 读取请求行
	//reqLine := make([]string,3)
	jlog.Debug("请求行:")
	if data, err := reader.ReadBytes('\n'); err == nil {

		//jlog.Debug(string(data[:len(data)-2]))
		reqLine = strings.Split(string(data[:len(data)-2]), " ")
		hm.reqMethod, hm.reqPath, hm.reqParams = hm.getInfoFromReqLine(reqLine)
		//tmpByte := bytes.Split(data,[]byte{' '})
		jlog.Debug(reqLine)
	}

	// 读取请求头
	reqHeaders = make(map[string]string)
	jlog.Debug("请求头:")
	for data, err := reader.ReadBytes('\n'); err == nil || err == io.EOF; data, err = reader.ReadBytes('\n') {
		if err == io.EOF {
			jlog.Fatal("报文格式错误", data)
			break
		}
		if len(data) == 2 {
			jlog.Debug("blank line")
			break
		} else {
			jlog.Debug(string(data[:len(data)-2]))
			// 保存请求头
			headerName := string(data[:bytes.IndexRune(data, ':')])
			reqHeaders[headerName] = strings.TrimLeft(string(data[bytes.IndexRune(data, ':')+1:len(data)-2]), " ")
		}
	}
	jlog.Debug(reqHeaders)
	hm.reqHeaders = reqHeaders
	hm.reqHost = hm.reqHeaders["Host"]
	// 读取请求体
	//var reqData []byte
	jlog.Debug("请求体:")
	for data, err := reader.ReadBytes('\n'); err == nil || err == io.EOF; data, err = reader.ReadBytes('\n') {
		reqData = data
		if err == io.EOF {
			break
		}
		//jlog.Debug(string(data[:len(data)-2]))
	}
	jlog.Debug(reqData)
	hm.reqData = reqData
	return reqLine, reqHeaders, reqData
}
