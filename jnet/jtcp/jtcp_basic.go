package jtcp

import (
	"github.com/chroblert/jgoutils/jlog"
	"github.com/google/gopacket/pcap"
	"net"
	"strings"
)



type IPMacName struct{
	Ipv4    []string
	Ipv6    []string
	Mac     string
	IfName  string
	IfIndex int
}


// 获取IP地址的类型
// return: 4,6,0
func GetIPType(ipStr string) int{
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return  0
	}
	for i := 0; i < len(ipStr); i++ {
		switch ipStr[i] {
		case '.':
			return  4
		case ':':
			return  6
		}
	}
	return  0
}

// 获取所有的网络设备
func GetNetWorks() []*localNetwork {
	tmpNetWorks := make([]*localNetwork,0)
	// find all devices
	devices,err := pcap.FindAllDevs()
	if err != nil{
		jlog.Fatal(err)
	}
	// 枚举所有的网络设备
	for _,v := range devices{
		if len(v.Addresses) == 0 {
			continue
		}
		tmp := new(localNetwork)
		for _,v2 := range v.Addresses{
			switch GetIPType(v2.IP.String()){
			case 4:
				tmp.localDevice = v.Name
				tmp.localIP = v2.IP.String()
				tmp.localMAC = GetMACByIP(v2.IP.String())
				if tmp.localMAC != ""{
					tmpNetWorks = append(tmpNetWorks,tmp)
				}
			case 6:
			}
		}
	}
	return tmpNetWorks
}

// 显示所有的网络设备，ip，Mac
func ShowNetworks(){
	for i,v := range GetNetWorks() {
		jlog.Println(i,v.localDevice,":",v.localIP,":",v.localMAC)
	}
}




// 根据IP获取mac
func GetMACByIP(ipv4 string) string{
	for _,v := range GetIPMACName(){
		if strings.Contains(strings.Join(v.Ipv4," "),ipv4){
			return v.Mac
		}
	}
	return ""
}

// 获取系统上所有的网卡，ip，Mac，IfName
// return: map[string]*IPMacName
func GetIPMACName() (ipMac []*IPMacName) {
	//ipMac := make(map[string]*IPMacName)
	netInterfaces, err := net.Interfaces()
	if err != nil {
		jlog.Printf("fail to get net interfaces: %v", err)
		return
	}
	// 获取电脑上所有的网卡，ip，Mac
	for _, netInterface := range netInterfaces {
		macAddr := netInterface.HardwareAddr.String()
		if len(macAddr) != 0 {
			ipAddrs,err := netInterface.Addrs()
			if err != nil{
				jlog.Error(err)
				continue
			}
			ipv4Addrs := []string{}
			ipv6Addrs := []string{}
			for _,v := range ipAddrs{
				ipStr := strings.Split(v.String(),"/")[0]
				if ipStr == "169.254.144.0"{
					break
				}
				switch GetIPType(ipStr){
				case 4:
					ipv4Addrs = append(ipv4Addrs,ipStr)
				case 6:
					ipv6Addrs = append(ipv6Addrs,ipStr)
				}
			}
			if len(ipv4Addrs) != 0 {
				ipMac = append(ipMac,&IPMacName{
					Ipv4:    ipv4Addrs,
					Ipv6:    ipv6Addrs,
					Mac:     macAddr,
					IfName:  netInterface.Name,
					IfIndex: netInterface.Index,
				})
			}
		}
	}
	return
}

// 获取一个可用的空闲端口
func GetFreePort(ipStr string) (uint16, error) {
	addr, err := net.ResolveTCPAddr("tcp", ipStr+":0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	jlog.Debug(uint16(l.Addr().(*net.TCPAddr).Port))
	return uint16(l.Addr().(*net.TCPAddr).Port), nil
}