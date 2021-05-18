package jhttp

import (
	"bufio"
	"bytes"
	"github.com/chroblert/JC-GoUtils/jasync"
	"github.com/chroblert/JC-GoUtils/jconv"
	"github.com/chroblert/JC-GoUtils/jlog"
	"github.com/chroblert/JC-GoUtils/jmath"
	"io"
	"os"
	"strconv"
)

func (hm *httpMsg) Intrude(isPrintAllStaus bool) {
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
		jasyncobj.Add(strconv.Itoa(i), singleIntruder, nil, newReqBytes, hm.isUseSSL)
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
	jasyncobj.Clean()
}

func singleIntruder(reqBytes []byte, isUseSSL bool) {
	hm := New()
	hm.InitWithBytes(reqBytes)
	hm.SetIsUseSSL(isUseSSL)
	hm.SetIsVerifySSL(false)
	hm.Repeat()
}
