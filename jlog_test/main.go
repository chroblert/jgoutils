package main

import (
	"bufio"
	"github.com/chroblert/jgoutils/jlog"
	"os"
	"strings"
	"time"
)

var (
	nlog *jlog.FishLogger
	alog *jlog.FishLogger
)

func main() {
	nlog = jlog.NewLogger(jlog.LogConfig{
		BufferSize:        2048,
		FlushInterval:     10 * time.Second,
		MaxStoreDays:      5,
		MaxSizePerLogFile: 512000000,
		LogCount:          5,
		LogFullPath:       "logs\\dishininormal-all.log",
		Lv:                jlog.DEBUG,
		UseConsole:        true,
		Verbose:           true,
		InitCreateNewLog:  false,
	})
	alog = jlog.NewLogger(jlog.LogConfig{
		BufferSize:        2048,
		FlushInterval:     10 * time.Second,
		MaxStoreDays:      5,
		MaxSizePerLogFile: 512000000,
		LogCount:          5,
		LogFullPath:       "logs\\dishiniall.log",
		Lv:                jlog.DEBUG,
		UseConsole:        true,
		Verbose:           true,
	})
	//jlog.SetLogFullPath("logs\\dishing.log")

	ReadLine("F:\\Data\\GO\\test210526\\logs\\dishininormal2020-2030.log")
	defer func() {
		jlog.Flush()
		alog.Flush()
		nlog.Flush()
	}()
}

func ReadLine(filename string) {
	f, err := os.Open(filename)
	if err != nil {
		jlog.Error(err)
		return
	}
	jlog.Debug("opened file")
	defer func() {
		f.Close()
		jlog.Debug("退出")

	}()
	r := bufio.NewReader(f)
	for {
		line, err := readLine(r)
		if err != nil {
			jlog.Error(err)
			break
		}
		nlog.NInfo(strings.Split(line, " | ")[6])
		// 对数据进行json unmashal
		//tmp := make(map[string]interface{})
		//err = json.Unmarshal([]byte(line),&tmp)
		//if err != nil {
		//	continue
		//}
		//// 判断host是否为ajax.quncrm.com；不为，则跳过
		//if tmp["__time__"] != nil  && tmp["host"] != nil && tmp["host"].(string) == "ajax.quncrm.com" {
		//	// 时间戳转换
		//	timestampstr := tmp["__time__"].(string)
		//	timeint, err := strconv.Atoi(timestampstr)
		//	var timestr string
		//	if err == nil {
		//		tm := time.Unix(int64(timeint), 0)
		//		timestr = tm.Format("2006-01-02 15:04:05")
		//	} else {
		//		jlog.Error(err)
		//		break
		//	}
		//	// 判断body里面是否有certificateNo;没有，则跳过
		//	if tmp["body"] != nil && strings.Contains(tmp["body"].(string), "certificateNo") {
		//		// 收集tmp["method"] tmp["host"] tmp["@formattedUrl"] tmp["body"] tmp["userAgent"] tmp["httpXForwardedFor"] tmp["referer"] tmp["__time__"]
		//		// 收集姓名，身份证，ip，信息
		//		alog.NInfof("%s | %s | %s | %s | %s | %s | %s | %s \n", timestr, tmp["method"], tmp["host"], tmp["@formattedUrl"], tmp["body"], tmp["userAgent"], tmp["httpXForwardedFor"], tmp["referer"])
		//	} else if tmp["__time__"] != nil && tmp["body"] != nil && tmp["@formattedUrl"] != nil && strings.Contains(tmp["@formattedUrl"].(string), "GET /:id/api/disneycommunity/merch/get-activity-config") {
		//		// 获取正常用户，判断tmp["@formattedUrl"]是否包含"GET /:id/api/disneycommunity/merch/get-activity-config"; 若无，则跳过（可疑人员）
		//		nlog.NInfof("%s | %s | %s | %s | %s | %s | %s | %s \n", timestr, tmp["method"], tmp["host"], tmp["@formattedUrl"], tmp["body"], tmp["userAgent"], tmp["httpXForwardedFor"], tmp["referer"])
		//	}
		//	//break
		//}

	}
}

func readLine(r *bufio.Reader) (string, error) {
	line, isprefix, err := r.ReadLine()
	for isprefix && err == nil {
		var bs []byte
		bs, isprefix, err = r.ReadLine()
		line = append(line, bs...)
	}
	return string(line), err
}
