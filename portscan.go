package main

import (
	"fmt"
	"github.com/chroblert/jgoutils/jasync"
	"github.com/chroblert/jgoutils/jlog"
	"github.com/chroblert/jgoutils/jnet/jparser"
	"github.com/chroblert/jgoutils/jnet/jtcp"
	"github.com/chroblert/jgoutils/jnet/jtcp/jcore"
	"strconv"
	"sync"
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
	if jtcpobj != nil{
		jcore.ShowNetworks()
		//jtcpobj.SetNetwork(2)
		jasyncobj := jasync.New()
		var wg = new(sync.WaitGroup)
		for _,v := range t{
			//jlog.Info(v)
			for _,v2 := range p{
				wg.Add(1)
				go func(v string,v2 int){
					jasyncobj.Add(v+":"+strconv.Itoa(v2),jtcpobj.SinglePortSYNScan,print,v,uint16(v2),"test")
					wg.Done()
				}(v,v2)
			}
		}
		wg.Wait()
		//jlog.Debug("sleep 100 second")
		//time.Sleep(100*time.Second)
		jasyncobj.Run(rate)
		//jtcpobj.RecvScanRes()

		jasyncobj.Wait()
		//jtcpobj.Test()
		jasyncobj.Clean()
		//jtcpobj.Test()
		//time.Sleep(3*time.Second)
		jtcpobj.CloseHandle()
	}
	return nil
}
