package jroute

import (
	"net"
)

type Route struct {
	Network net.IP
	Mask    byte
}

func (r *Route) GetIpv4Mask() net.IP {
	return net.IP(net.CIDRMask(int(r.Mask), 32))
}

type RouteTable interface {
	// 增加 Net 路由表
	// Net 路由表指的是走非 VPN 的路由表
	// 即 VPN 白名单。
	// 内部根据默认路由确定 WAN 网关地址。
	AddNetRoutes(routes []Route) error

	// 增加 VPN 路由表
	// 走 VPN 网络的路由表
	// 内部会将 0.0.0.0/0 拆分为两个，防止干扰造成找不到默认路由表的问题
	// 内部根据提供的本地接口确定
	AddVpnRoutes(routes []Route, network, mask, gIp net.IP) error

	// 删除之前添加的路由表
	//DelRoutes(Routes) error

	// 清洗功能
	// 清洗路由表，删除所有非本地接口、非默认网关的路由条目、跃点数为X的路由条目...
	// 感觉这个实现并不好。
	ResetRoute() error

	// 获取接口IP的网关IP
	GetGatewayByDstIP(ifIPStr string) (string,error)
}
