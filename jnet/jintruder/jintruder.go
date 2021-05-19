package jintruder

import (
	"bufio"
	"bytes"
	"github.com/chroblert/jgoutils/jconv"
	"github.com/chroblert/jgoutils/jlog"
	"github.com/chroblert/jgoutils/jnet/jhttp"
	"io"
	"io/ioutil"
	"os"
)

type httpMsg struct {
}

func init() {
	intruder("./req.txt", "c:\\data\\test1.txt", "c:\\data\\test2.txt")
}

func intruder(filename string, wordfile ...string) {
	reqbytes, _ := ioutil.ReadFile(filename)
	stringss := [][]string{}
	for k, v := range wordfile {
		jlog.Info(k, v)
		rd, err := os.Open(v)
		if err != nil {
			jlog.Error(err)
		}
		reader := bufio.NewReader(rd)
		lines := []string{}
		for {
			if line, _, err := reader.ReadLine(); err == nil || err == io.EOF {
				jlog.Info(string(line))
				lines = append(lines, string(line))
				if err == io.EOF {
					break
				}
			}
		}
		stringss = append(stringss, lines)
	}

	wordlists := Dikaer(jconv.ConvertStringssToInterfacess(stringss))
	for _, wordtuple := range wordlists {
		idx := 0
		for {
			if !bytes.ContainsAny(reqbytes, "*") {
				break
			}
			reqbytes = bytes.Replace(reqbytes, []byte("*"), []byte(wordtuple[idx].(string)), 1)
			idx++
			if len(wordtuple) < idx {
				jlog.Error("error,字典个数少于标识的个数")
			}
		}
		//jlog.Info(string(reqbytes))
		jhttpobj := jhttp.New()
		jhttpobj.InitWithBytes(reqbytes)
		jlog.Info(jhttpobj.Repeat())
	}

}

// 获取多个数组的排列组合，即笛卡尔积
func Dikaer(sets [][]interface{}) [][]interface{} {
	lens := func(i int) int { return len(sets[i]) }
	product := [][]interface{}{}
	for ix := make([]int, len(sets)); ix[0] < lens(0); nextIndex(ix, lens) {
		var r []interface{}
		for j, k := range ix {
			r = append(r, sets[j][k])
		}
		product = append(product, r)
	}
	return product
}

func nextIndex(ix []int, lens func(i int) int) {
	for j := len(ix) - 1; j >= 0; j-- {
		ix[j]++
		if j == 0 || ix[j] < lens(j) {
			return
		}
		ix[j] = 0
	}
}
