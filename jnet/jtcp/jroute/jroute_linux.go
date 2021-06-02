// +build linux

package jroute

import "github.com/chroblert/jgoutils/jlog"

type linuxRouteTable struct{

}
func NewRouteTable() RouteTable {

	return &linuxRouteTable{}
}

func (lrt *linuxRouteTable)GetGatewayByDstIP(ifIPStr string) (string,error){
	router,err := routing.New()
	if err != nil{
		jlog.Fatal(err)
	}
	iface, gw, src, err := router.Route(net.ParseIP(ifIPStr))
	if err != nil {
		jlog.Error(err)
		return "",err
	}
	return gw.String(),nil
}
