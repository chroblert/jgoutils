package main

import (
	"github.com/chroblert/jgoutils/jlog"
	"github.com/chroblert/jgoutils/jnet/jparser"
)

func portScan(ipStr string,portStr string,rate int){
	t := jparser.ParseIPStr(ipStr)
	for _,v := range t{
		jlog.Info(v)
	}
	p := jparser.ParsePortStr(portStr)
	for _,v := range p{
		jlog.Info(v)
	}

}
