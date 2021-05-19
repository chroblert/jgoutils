package jhttp

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/chroblert/jgoutils/jasync"
	"github.com/chroblert/jgoutils/jconv"
	"github.com/chroblert/jgoutils/jlog"
	"github.com/chroblert/jgoutils/jmath"
	"io"
	"os"
	"strconv"
)

func (hm *httpMsg) Intrude(isPrintAllStaus bool, printWithFilter func(statuscode int, headers map[string][]string, body []byte, err error)) map[string][]interface{} {
	if len(hm.intruData.wordFiles) < 1 {
		jlog.Fatal("请设置至少一个字典文件")
	}
	//reqbytes, _ := ioutil.ReadFile(filename)
	stringss := [][]string{}
	for k, v := range hm.intruData.wordFiles {
		jlog.Debug("打开字典文件:", k, v)
		rd, err := os.Open(v)
		if err != nil {
			jlog.Fatal(err)
		}
		reader := bufio.NewReader(rd)
		lines := []string{}
		for {
			if line, _, err := reader.ReadLine(); err == nil || err == io.EOF {
				//jlog.Info(string(line))
				if err == io.EOF {
					if string(line) != "" {
						lines = append(lines, string(line))
					}
					break
				}
				lines = append(lines, string(line))
			}
		}
		stringss = append(stringss, lines)
	}
	newReqBytes := make([]byte, 0)
	wordlists := jmath.Dikaer(jconv.ConvertStringssToInterfacess(stringss))
	jasyncobj := jasync.New()
	for i, wordtuple := range wordlists {
		newReqBytes = hm.intruData.reqBytes
		idx := 0
		for {
			if !bytes.ContainsAny(newReqBytes, "*") {
				break
			}
			newReqBytes = bytes.Replace(newReqBytes, []byte("*"), []byte(wordtuple[idx].(string)), 1)
			idx++
			if len(wordtuple) < idx {
				jlog.Error("error,字典个数少于标识的个数")
			}
		}
		jasyncobj.Add(strconv.Itoa(i), singleIntruder, printWithFilter, newReqBytes, hm.isUseSSL, hm.getProxy())
		////jlog.Info(string(reqbytes))
		//jhttpobj := jhttp.New()
		//jhttpobj.InitWithBytes(reqbytes)
		//jlog.Info(jhttpobj.Repeat())
		//hm.InitWithBytes(newReqBytes)
		//hm.Repeat()
	}
	if jasyncobj.GetTotal() > 0 {
		jasyncobj.Run()
		jasyncobj.Wait()
		if isPrintAllStaus {
			jasyncobj.GetStatus("", false)
		}
	}
	result := jasyncobj.GetTasksResult()
	jasyncobj.Clean()
	return result
}

func singleIntruder(reqBytes []byte, isUseSSL bool, proxy string) (statuscode int, headers map[string][]string, body []byte, err error) {
	hm2 := New()
	hm2.InitWithBytes(reqBytes)
	hm2.SetIsUseSSL(isUseSSL)
	hm2.SetIsVerifySSL(false)
	hm2.SetProxy(proxy)
	tmp := hm2.Repeat()
	hm2.Clean()
	return tmp["0"][0].(int), tmp["0"][1].(map[string][]string), tmp["0"][2].([]byte), fmt.Errorf("%v", tmp["0"][3])
}
