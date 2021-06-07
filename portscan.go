package main

import (
	"fmt"
	"github.com/chroblert/jgoutils/jlog"
	"github.com/chroblert/jgoutils/jnet/jparser"
	"github.com/chroblert/jgoutils/jnet/jtcp"
	"github.com/chroblert/jgoutils/jnet/jtcp/jcore"
	"github.com/panjf2000/ants/v2"
	"github.com/petermattis/goid"
	"strconv"
	"sync"
	"time"
)

func portScan(ipStr string,portStr string,rate int) error{
	t := jparser.ParseIPStr(ipStr)
	if len(t) == 0{
		jlog.Error("nil ip")
		return fmt.Errorf("nil ip")
	}
	p := jparser.ParsePortStr(portStr)
	if len(p) == 0{
		jlog.Error("nil port")
		return fmt.Errorf("nil port")
	}
	jtcpobj := jtcp.New()
	go jtcpobj.RecvScanRes("","","")
	var wg = new(sync.WaitGroup)
	// rate Mb
	//rate = rate*1024/66
	jlog.Warn(goid.Get(),": send packet rate:",rate)
	poo,_ := ants.NewPoolWithFunc(rate, func(i interface{}) {
		tmp := i.([]string)
		port,_ := strconv.Atoi(tmp[1])
		jtcpobj.SinglePortSYNScan(tmp[0],uint16(port))
		wg.Done()
	})
	if jtcpobj != nil{
		jcore.ShowNetworks()
		//jtcpobj.SetNetwork(2)
		//jasyncobj := jasync.New()
		for _,v := range t{
			//jlog.Info(v)
			for _,v2 := range p{
				wg.Add(1)
				tmp := make([]string,2)
				tmp[0] = v
				tmp[1] = strconv.Itoa(v2)
				//go func(v string,v2 int){
				//	jasyncobj.Add(v+":"+strconv.Itoa(v2),jtcpobj.SinglePortSYNScan,nil,v,uint16(v2))
				//	wg.Done()
				//}(v,v2)
				go poo.Invoke(tmp)
			}
		}
		wg.Wait()
		//jasyncobj.Run(rate)
		//jasyncobj.Wait()
		//jasyncobj.Clean()
		time.Sleep(3*time.Second)
		jtcpobj.CloseHandle()
	}
	return nil
}
