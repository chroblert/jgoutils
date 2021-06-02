package main

import (
	_ "github.com/chroblert/jgoutils/jconfig"
	"github.com/chroblert/jgoutils/jlog"
	"github.com/chroblert/jgoutils/jnet/jtcp"
	"strings"
	//_ "github.com/chroblert/jgoutils/jtest"
	//_ "github.com/chroblert/jgoutils/jnet/jintruder"
)

const (
	HttpProxy = "http://192.168.50.2:8080"
	//HttpProxy = "http://192.168.30.109:10809"
	SocksProxy = "socks5://192.168.30.109:10808"
)

func Print(statusCode int, headers map[string][]string, body []byte, err error) {
	//fmt.Println(strings.TrimSpace(string(body)))
	jlog.Info("状态码：", statusCode, strings.TrimSpace(string(body)))
}

//func print(test int){
//	fmt.Println(test)
//}

func main() {
	//start := time.Now()
	//ch := make(chan jasyncrequests.Asyncresponse)
	//for i:= 0;i<10;i++{
	//	requrl:= fmt.Sprintf("http://myip.ipip.net/?q=%d",i)
	//	go jasyncrequests.Fetch(fmt.Sprintf("%d",i),requrl, ch) // start a goroutine
	//}
	//for i:= 0;i<10;i++{
	//	rst := <-ch
	//	fmt.Println(rst.Key,rst.TimeElapsed,rst.Statuscode,strings.TrimSpace(string(rst.Body))) // receive from channel ch
	//}
	//fmt.Printf("%.2fs elapsed\n", time.Since(start).Seconds())

	//// ====== jasync测试
	//asyncobj := jasync.New()
	////asyncobj.Add("request1", jrequests.Get,print,"http://myip.ipip.net/?q=1",jrequests.SetProxy(HttpProxy))
	////asyncobj.Add("request2", jrequests.Get,print,"http://myip.ipip.net/?q=2",jrequests.SetProxy(HttpProxy))
	//for i := 0;i<1;i++{
	//	asyncobj.Add("request"+strconv.Itoa(i), jrequests.Get,print,"http://myip.ipip.net/?q="+strconv.Itoa(i),jrequests.SetProxy(HttpProxy))
	//}
	//
	//// 执行
	////if chans,ok := asyncobj.Run();ok{
	////	// 将数据从通道中取回,取回的值是一个map[string]interface{}类型,key为async.Add()时添加的logo,interface{}为该logo回调函数返回的结果
	////	res := <-chans
	////	//fmt.Println(len(res))
	////	// 这里最好判断下是否所有的异步请求都已经执行成功
	////	if len(res) == asyncobj.GetTotal() {
	////
	////	} else {
	////		log.Println("jasync not execution all task")
	////	}
	////}
	//asyncobj.Run()
	////fmt.Println(len(chans))
	////fmt.Println(<-chans)
	////fmt.Println(len(chans))
	//time.Sleep(1*time.Second)
	//asyncobj.GetStatus("false",false)
	//asyncobj.Wait()
	//asyncobj.GetStatus("false",false)
	//// 清除掉本次操作的所有数据,方便后续继续使用async对象
	//asyncobj.Clean()

	// ==== jlog测试
	jlog.SetLevel(jlog.INFO)
	//jlog.Debug("debug")
	//jlog.Info("info")
	//jlog.Warn("warn")
	//jlog.Error("error")
	//jlog.Fatal("fatal")
	defer jlog.Flush()

	// === jhttp测试
	//fileName := "req.txt"
	//jhttpobj := jhttp.New()
	////jhttpobj.SetURL("https://test.com/query/string?q1=ddd&q2=xxx&q3[1]=xxxdd&q3[2]=qe2")
	//jhttpobj.InitWithFile(fileName)
	//jhttpobj.SetWordfiles("c:\\data\\test1.txt")
	//jhttpobj.SetIsUseSSL(false)
	//jhttpobj.SetProxy("http://192.168.30.109:10809")
	//jhttpobj.SetIsVerifySSL(false)
	////jhttpobj.SetReqMethod("GET")
	//jhttpobj.Intrude(false, func(statuscode int, headers map[string][]string, body []byte, err error) {
	//	jlog.Info(statuscode, err)
	//})
	//jlog.Info(tmp)
	//jhttpobj.Repeat(5)
	//jhttpobj.InitWithFile(fileName)
	//jhttpobj.SetWordfiles("c:\\data\\test1.txt")
	//jasyncobj := jasync.New()
	//for i := 0; i < 20; i++ {
	//	jasyncobj.Add(strconv.Itoa(i), jhttpobj.Repeat, Print)
	//}
	//jasyncobj.Run()
	////jhttpobj.Repeat()
	//jasyncobj.Wait()
	////jasyncobj.GetStatus("",false)
	//jasyncobj.Clean()
	//jhttpobj.Repeat(10)
	//jhttpobj.Intrude(true)
	jlog.Info("over")
	//jlog.Println("test")

	// ==== jconfig测试
	//code, _, _, _ := jrequests.Get("https://www.baidu.com")
	//jlog.Debug(code)
	//jlog.Debug(string(body))

	//jhttpobj := jhttp.New()
	//jhttpobj.InitWithBytes([]byte("GET / http/1.1\r\nHost: baidu.com\r\nContent-Length: 18\r\n\r\ntest"))
	//jhttpobj.Repeat()

	// grumble2测试
	//var app = grumble.New(&grumble.Config{
	//	Name:                  "app",
	//	Description:           "app test",
	//	Flags:                 nil,
	//	HistoryFile:           "",
	//	HistoryLimit:          0,
	//	NoColor:               false,
	//	Prompt:                "app >> ",
	//	PromptColor:           nil,
	//	MultiPrompt:           "",
	//	MultiPromptColor:      nil,
	//	ASCIILogoColor:        nil,
	//	ErrorColor:            nil,
	//	HelpHeadlineUnderline: false,
	//	HelpSubCommands:       false,
	//	HelpHeadlineColor:     nil,
	//	CurrentCommand:        "app",
	//})
	//app.AddCommand(&grumble.Command{
	//	Name:      "test",
	//	FullPath: "testxxx/test",
	//	Aliases:   nil,
	//	Help:      "this is a test command",
	//	LongHelp:  "",
	//	HelpGroup: "testGroup",
	//	Usage:     "test arg flag",
	//	Flags: func(f *grumble.Flags) {
	//		f.String("s","stest","","string test")
	//		f.Bool("b","isss",false,"bool test")
	//		f.StringSlice("c","slice",[]string{},"string slice test")
	//	},
	//	Args:      nil,
	//	Run: func(c *grumble.Context) error {
	//		jlog.Println(c.Flags.String("stest"))
	//		jlog.Println(c.Flags.Bool("isss"))
	//		jlog.Println(c.Flags.StringSlice("slice"))
	//		return nil
	//	},
	//	Completer: nil,
	//})
	//app.Run()

	// jtcp测试
	jtcp.ShowNetworks()
}
