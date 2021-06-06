// +build windows

package jroute

import (
	"bytes"
	"fmt"
	"github.com/chroblert/jgoutils/jlog"
	"github.com/chroblert/jgoutils/jnet/jtcp/jcore"
	"net"
	"syscall"
	"unsafe"
)
// JC0o0l Comment: 参考https://github.com/GameXG/gonet/tree/master/route
/*
对于路由表，预期的方法是：
查询 0.0.0.0/0 获得原始默认路由
然后为 vpn 服务器添加默认路由
之后就根据需要下发vpn路由完事。
对于0.0.0.0/0 vpn 路由，可以尝试更低的跃点数，也可以尝试分为2个。
重新连接时可以删除vpn接口的所有非链路路由表。
路由表格式：
目标网络 uint32   掩码位数 byte低6位  vpn/默认网关  byte 高1位
*/

// 太低的值添加路由时会返回 106 错误
const routeMetric = 93

type routeTable struct {
	iphlpapi             *syscall.LazyDLL
	getIpForwardTable    *syscall.LazyProc
	createIpForwardEntry *syscall.LazyProc
	deleteIpForwardEntry *syscall.LazyProc
}

type RouteRow struct {
	ForwardDest      [4]byte //目标网络
	ForwardMask      [4]byte //掩码
	ForwardPolicy    uint32  //ForwardPolicy:0x0
	ForwardNextHop   [4]byte //网关
	ForwardIfIndex   uint32  // 网卡索引 id
	ForwardType      uint32  //3 本地接口  4 远端接口
	ForwardProto     uint32  //3静态路由 2本地接口 5EGP网关
	ForwardAge       uint32  //存在时间 秒
	ForwardNextHopAS uint32  //下一跳自治域号码 0
	ForwardMetric1   uint32  //度量衡(跃点数)，根据 ForwardProto 不同意义不同。
	ForwardMetric2   uint32
	ForwardMetric3   uint32
	ForwardMetric4   uint32
	ForwardMetric5   uint32
}

func (rr *RouteRow) GetForwardDest() net.IP {
	return net.IP(rr.ForwardDest[:])
}
func (rr *RouteRow) GetForwardMask() net.IP {
	return net.IP(rr.ForwardMask[:])
}
func (rr *RouteRow) GetForwardNextHop() net.IP {
	return net.IP(rr.ForwardNextHop[:])
}

func NewRouteTable() RouteTable {
	iphlpapi := syscall.NewLazyDLL("iphlpapi.dll")
	getIpForwardTable := iphlpapi.NewProc("GetIpForwardTable")
	createIpForwardEntry := iphlpapi.NewProc("CreateIpForwardEntry")
	deleteIpForwardEntry := iphlpapi.NewProc("DeleteIpForwardEntry")

	return &routeTable{
		iphlpapi:             iphlpapi,
		getIpForwardTable:    getIpForwardTable,
		createIpForwardEntry: createIpForwardEntry,
		deleteIpForwardEntry: deleteIpForwardEntry,
	}
}

func (rt *routeTable) getRoutes() ([]RouteRow, error) {
	buf := make([]byte, 4+unsafe.Sizeof(RouteRow{}))
	buf_len := uint32(len(buf))

	rt.getIpForwardTable.Call(uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&buf_len)), 0)

	var r1 uintptr
	for i := 0; i < 5; i++ {
		buf = make([]byte, buf_len)
		r1, _, _ = rt.getIpForwardTable.Call(uintptr(unsafe.Pointer(&buf[0])),
			uintptr(unsafe.Pointer(&buf_len)), 0)
		if r1 == 122 {
			continue
		}
		break
	}

	if r1 != 0 {
		return nil, fmt.Errorf("Failed to get the routing table, return value：%v", r1)
	}

	num := *(*uint32)(unsafe.Pointer(&buf[0]))
	routes := make([]RouteRow, num)
	sr := uintptr(unsafe.Pointer(&buf[0])) + unsafe.Sizeof(num)
	rowSize := unsafe.Sizeof(RouteRow{})

	// 安全检查
	if len(buf) < int((unsafe.Sizeof(num) + rowSize*uintptr(num))) {
		return nil, fmt.Errorf("System error: GetIpForwardTable returns the number is too long, beyond the buffer。")
	}

	for i := uint32(0); i < num; i++ {
		routes[i] = *((*RouteRow)(unsafe.Pointer(sr + (rowSize * uintptr(i)))))
	}

	return routes, nil
}

func (rt *routeTable) addRoute(rr *RouteRow) error {
	fmt.Printf("准备添加路由表 %#v ...", rr)
	r1, _, err := rt.createIpForwardEntry.Call(uintptr(unsafe.Pointer(rr)))
	fmt.Printf("r:%v,err:%v", r1, err)
	if r1 == 5010 {
		// 指定的路由条目已存在
		return nil
	} else if r1 != 0 {
		return fmt.Errorf("Add routing table%#v error, return value：%v ,err:%v", rr, r1, err)
	}
	return nil
}

func (rt *routeTable) delRoute(rr *RouteRow) error {
	r1, _, err := rt.deleteIpForwardEntry.Call(uintptr(unsafe.Pointer(rr)))
	if r1 != 0 {
		return fmt.Errorf("Delete routing table%#v error, return value：%v ,err:%v", rr, r1, err)
	}
	return nil
}

func (rt *routeTable) AddNetRoutes(routes []Route) error {
	rs, err := rt.getRoutes()
	if err != nil {
		return err
	}

	var defaultRoute *RouteRow
	for _, r := range rs {
		if r.ForwardType == 4 && r.ForwardProto == 3 &&
			bytes.Equal([]byte{0, 0, 0, 0}, []byte(r.GetForwardDest())) &&
			bytes.Equal([]byte{0, 0, 0, 0}, []byte(r.GetForwardMask())) &&
			r.ForwardMetric1 != routeMetric {
			if defaultRoute != nil {
				if defaultRoute.ForwardMetric1 < r.ForwardMetric1 {
					continue
				}
			}
			t := r
			defaultRoute = &t
		}
	}

	if defaultRoute == nil {

		return fmt.Errorf("Default gateway not found.")
	}

	defaultRoute.ForwardMetric1 = routeMetric
	for _, r := range routes {
		if n := copy(defaultRoute.ForwardDest[:], []byte(r.Network.To4())); n != 4 {
			return fmt.Errorf("internal error,copy(defaultRoute.ForwardDest[:], []byte(r.Ip.To4())) return %v != 4", n)
		}
		if n := copy(defaultRoute.ForwardMask[:], []byte(r.GetIpv4Mask())); n != 4 {
			fmt.Println(r.GetIpv4Mask())
			return fmt.Errorf("internal error,copy(defaultRoute.ForwardMask[:], []byte(r.GetIpv4Mask())) return %v != 4", n)
		}
		if err := rt.addRoute(defaultRoute); err != nil {
			return fmt.Errorf("Add routing table failed,%v", err)
		}
	}

	return nil
}
func (rt *routeTable) AddVpnRoutes(routes []Route, network, mask, gIp net.IP) error {
	networkv4 := network.To4()
	maskv4 := mask.To4()
	gIpv4 := gIp.To4()
	if len(networkv4) != 4 || len(maskv4) != 4 || len(gIpv4) != 4 {
		return fmt.Errorf("network:%v,mask:%v,gIp:%v 不是 Ipv4地址。", networkv4, maskv4, gIpv4)
	}

	rs, err := rt.getRoutes()
	if err != nil {
		return err
	}

	var vpnRoute *RouteRow
	for _, r := range rs {
		jlog.Debug("ForwardType:", r.ForwardType, "ForwardProto:", r.ForwardProto, "network:", r.GetForwardDest(), "mask:", r.GetForwardMask(), "g:", r.GetForwardNextHop())
		if r.ForwardType == 3 && //r.ForwardProto == 3 &&
			r.GetForwardDest().Equal(networkv4) && r.GetForwardMask().Equal(maskv4) {
			t := r
			vpnRoute = &t
			break
		}
	}

	if vpnRoute == nil {
		return fmt.Errorf("VPN route not found.")
	}

	vpnRoute.ForwardMetric1 = routeMetric
	vpnRoute.ForwardType = 4
	vpnRoute.ForwardProto = 3

	if n := copy(vpnRoute.ForwardNextHop[:], []byte(gIpv4)); n != 4 {
		return fmt.Errorf("内部错误，copy 返回值 %v != 4", n)
	}

	for _, r := range routes {
		if r.Network.Equal(net.IPv4(0, 0, 0, 0).To4()) && r.Mask == 0 {
			// 对于默认路由，需要特殊处理下 做法是彩粉成为多个网络地址。
			// 好处是不会因为优先级的问题使得下发的默认网关无效
			// 坏处是使得windows的默认网管故障自动切换策略失效，不过在 vpn 终端后，相关路由应该也就失效了，所以应该没影响。
			if n := copy(vpnRoute.ForwardDest[:], []byte{128, 0, 0, 0}); n != 4 {
				return fmt.Errorf("内部错误，copy(defaultRoute.ForwardDest[:], []byte(r.Ip.To4())) 返回值 %v != 4", n)
			}
			if n := copy(vpnRoute.ForwardMask[:], []byte{128, 0, 0, 0}); n != 4 {
				return fmt.Errorf("内部错误，copy(defaultRoute.ForwardMask[:], []byte(r.GetIpv4Mask())) 返回值 %v != 4", n)
			}
			if err := rt.addRoute(vpnRoute); err != nil {
				return fmt.Errorf("添加路由表失败，%v", err)
			}

			if n := copy(vpnRoute.ForwardDest[:], []byte{0, 0, 0, 0}); n != 4 {
				return fmt.Errorf("内部错误，copy(defaultRoute.ForwardDest[:], []byte(r.Ip.To4())) 返回值 %v != 4", n)
			}
			if n := copy(vpnRoute.ForwardMask[:], []byte{128, 0, 0, 0}); n != 4 {
				return fmt.Errorf("内部错误，copy(defaultRoute.ForwardMask[:], []byte(r.GetIpv4Mask())) 返回值 %v != 4", n)
			}
			if err := rt.addRoute(vpnRoute); err != nil {
				return fmt.Errorf("Add routing table failed，%v", err)
			}
			continue
		}

		if n := copy(vpnRoute.ForwardDest[:], []byte(r.Network.To4())); n != 4 {
			return fmt.Errorf("内部错误，copy 返回值 %v != 4", n)
		}
		if n := copy(vpnRoute.ForwardMask[:], []byte(r.GetIpv4Mask())); n != 4 {
			return fmt.Errorf("内部错误，copy 返回值 %v != 4", n)
		}
		if err := rt.addRoute(vpnRoute); err != nil {
			return fmt.Errorf("Add routing table failed，%v", err)
		}
	}

	return nil
}

func (rt *routeTable) ResetRoute() error {
	rs, err := rt.getRoutes()
	if err != nil {
		return err
	}

	for _, r := range rs {
		if r.ForwardType == 4 && r.ForwardProto == 3 && r.ForwardMetric1 == routeMetric {
			if err := rt.delRoute(&r); err != nil {
				return err
			}
		}
	}
	return nil
}


// 210602: JC0o0l add
// 获取接口IP的网关IP
func (rt *routeTable)GetGatewayByIfIP(ifIPStr string) (string,error){
	ipMacName := jcore.GetIPMACName()
	sucFlag := false
	// 获取接口IP的接口索引
	var ifIndex int
loop:
	for _,v := range ipMacName{
		for _,v2 := range v.Ipv4{
			if v2 == ifIPStr{
				ifIndex = v.IfIndex
				//jlog.Println(v.Ipv4,v.IfName,v.IfIndex)
				sucFlag = true
				break loop
			}
		}
	}
	// 判断是否成功获取到接口IP的索引
	if !sucFlag{
		jlog.Errorf("%v 不是接口IP\n",ifIPStr)
		return "",fmt.Errorf("%v 不是接口IP\n",ifIPStr)
	}
	rs,err := rt.getRoutes()
	if err != nil{
		jlog.Fatal(err)
	}
	for _,r := range rs{
		//jlog.Debug("ForwardType:", r.ForwardType, "ForwardProto:", r.ForwardProto, "network:", r.GetForwardDest(), "mask:", r.GetForwardMask(), "g:", r.GetForwardNextHop(),"ifIndex:",r.ForwardIfIndex)
		if uint32(ifIndex) == r.ForwardIfIndex{
			//jlog.Println(net.IPv4(r.ForwardNextHop[0],r.ForwardNextHop[1],r.ForwardNextHop[2],r.ForwardNextHop[3]).String())
			return net.IPv4(r.ForwardNextHop[0],r.ForwardNextHop[1],r.ForwardNextHop[2],r.ForwardNextHop[3]).String(),nil
			break
		}
	}
	return "",fmt.Errorf("没有找到接口IP对应的网关IP")
}