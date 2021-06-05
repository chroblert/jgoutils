// +build linux

package jroute

import (
	"github.com/chroblert/jgoutils/jlog"
	"github.com/chroblert/jgoutils/jthirdutil/github.com/jackpal/gateway"
)

type linuxRouteTable struct{

}
func NewRouteTable() RouteTable {

	return &linuxRouteTable{}
}


// get gateway ip by interface ip
func (lrt *linuxRouteTable) GetGatewayByIfIP(ifIPStr string) (string,error){

	ip,err := gateway.GetGatewayByIfIP(ifIPStr)
	if err != nil{
		jlog.Error(err)
		return "",err
	}
	//jlog.Debug(ip.String())
	return ip.String(),nil
	//
	//jlog.Debug(gateway.DiscoverGateway())
	//
	//router,err := routing.New()
	//if err != nil{
	//	//jlog.Fatal(err)
	//	return "",err
	//}
	////dstIPStr = "101.132.112.169"
	//jlog.Debug("ifIPStr:", ifIPStr)
	//iface, gw, src, err := router.Route(net.ParseIP(ifIPStr))
	//if err != nil {
	//	jlog.Error(err)
	//	return "",err
	//}
	//jlog.Debug(iface.Name,gw,src.String())
	//return gw.String(),nil
}
