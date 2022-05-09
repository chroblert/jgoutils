package jhttp

import (
	"bytes"
	"strings"
)

func (hm *httpMsg) parseFromBytes(reqMsg []byte) (reqLine []string, reqHeaders map[string]string, reqData []byte) {
	if !CheckReqMsgIsValid(reqMsg) {
		jHttpLog.Error("报文格式不对")
	}
	reqHeaders = make(map[string]string)
	for i, line := range bytes.Split(reqMsg, []byte("\r\n")) {
		// 第一行reqLine
		if i == 0 {
			reqLine = strings.Split(string(line), " ")
			hm.reqMethod, hm.reqPath, hm.reqParams = hm.getInfoFromReqLine(reqLine)
		} else if i == len(bytes.Split(reqMsg, []byte("\r\n")))-1 {
			reqData = line
			hm.reqData = line
		} else if string(line) != "" {
			idx := strings.Index(string(line), ":")
			reqHeaders[string(line)[:idx]] = strings.TrimLeft(string(line)[idx+1:], " ")
		}
	}
	hm.reqHeaders = reqHeaders
	hm.reqHost = hm.reqHeaders["Host"]

	return reqLine, reqHeaders, reqData
}

func CheckReqMsgIsValid(reqMsg []byte) bool {
	// 判断报文是否至少含有一行空格
	if !bytes.Contains(reqMsg, []byte("\r\n\r\n")) {
		jHttpLog.Error("至少包含一行空行")
		return false
	}
	// 判断第一行是否有两个空格
	idx := bytes.IndexRune(reqMsg, '\r')
	if len(bytes.Split(reqMsg[:idx], []byte(" "))) != 3 {
		jHttpLog.Error("首行应有两个空行")
		return false
	}
	return true
}
