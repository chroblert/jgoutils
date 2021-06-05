package jtcp

import (
	"fmt"
	"github.com/chroblert/jgoutils/jlog"
	"github.com/chroblert/jgoutils/jnet/jtcp/jcore"
	"github.com/chroblert/jgoutils/jnet/jtcp/jroute"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type localNetwork struct{
	localIP string
	localMAC string
	localDevice string
	//localPort string
}

type remoteNetwork struct{
	remoteIP string
	remotePort string
	remoteMAC string
}

type tcpMsg struct{
	localNetworkInst *jcore.LocalNetwork
	snapshot_len int32
	promiscuous bool
	timeout time.Duration

	handle       *pcap.Handle
	buffer       gopacket.SerializeBuffer
	options      gopacket.SerializeOptions

	remoteMAC string
	//portScanRes map[string]string // localIP:localPort-remoteIP:remotePort,status
	portScanRes *sync.Map // localIP:localPort-remoteIP:remotePort,status
	mu *sync.RWMutex
	//portScanTasks map[string]
	portScanTasks *sync.Map

}

func New() *tcpMsg {
	tmp := jcore.GetNetWorks()
	tm := &tcpMsg{
		localNetworkInst: tmp[0],
		snapshot_len:     1024,
		promiscuous:      false,
		timeout:          5 * time.Second,
		handle: &pcap.Handle{},
		buffer: gopacket.NewSerializeBuffer(),
		options: gopacket.SerializeOptions{
			FixLengths:       true,
			ComputeChecksums: true,
		},
		//portScanRes: make(map[string]string),
		portScanRes: &sync.Map{},
		mu: new(sync.RWMutex),
		//portScanTasks: make(map[string]int),
		portScanTasks: &sync.Map{},
	}
	for i := 0; i < len(tmp);i++{
		if err := tm.SetNetwork(i); err == nil{
			break
		}else if i == len(tmp) -1{
			jlog.Error("获取网关的MAC地址失败")
			return nil
		}
	}
	//// 获取路由表
	//rt := jroute.NewRouteTable()
	//// 从路由表中获取接口IP的网关IP
	//gwIPStr,err := rt.GetGatewayByIfIP(tm.localNetworkInst.LocalIP)
	//if err != nil{
	//	jlog.Error(err)
	//	return nil
	//}
	//tm.SetNetwork(0)
	//// 获取网关IP的MAC地址
	//gwMacStr,err := tm.GetHWAddr(gwIPStr)
	//if err != nil{ // 从接口0没有成功获取到网关IP的mac地址
	//	jlog.Error(err)
	//	for i := 1;i<len(tmp);i++{
	//		tm.SetNetwork(i)
	//		gwIPStr,err := rt.GetGatewayByIfIP(tm.localNetworkInst.LocalIP)
	//		if err != nil{
	//			jlog.Error(err)
	//			return nil
	//		}
	//		gwMacStr,err := tm.GetHWAddr(gwIPStr)
	//		if err == nil{
	//			tm.remoteMAC = gwMacStr
	//			break
	//		}
	//	}
	//}
	//tm.remoteMAC = gwMacStr
	return tm
}

func (p *tcpMsg)SetNetwork(id int) error{
	tmp := jcore.GetNetWorks()
	if len(tmp) <= id{
		return fmt.Errorf("id should less than number of interface. 0=< id < len(interface)")
	}
	// 设置本地网络
	p.localNetworkInst = tmp[id]
	// 关闭之前的handle
	if p.handle != nil{
		p.handle.Close()
	}
	var err error
	p.handle, err = pcap.OpenLive(p.localNetworkInst.LocalDevice, p.snapshot_len, p.promiscuous, p.timeout)
	if err != nil {
		jlog.Error(err)
		return err
	}
	// 获取该接口的网关的IP
	// 获取路由表
	rt := jroute.NewRouteTable()
	// 从路由表中获取接口IP的网关IP
	gwIPStr,err := rt.GetGatewayByIfIP(p.localNetworkInst.LocalIP)
	if err != nil{
		jlog.Error(err)
		return err
	}
	// 使用arp获取网关IP的mac地址
	gwMacStr,err := p.GetHWAddr(gwIPStr)
	if err != nil{
		return err
	}
	p.remoteMAC = gwMacStr
	return nil
}

func (p *tcpMsg)CloseHandle(){
	p.handle.Close()
}

//func (p *tcpMsg)SetPortScanRes(key string,val string){
//	if val != "open" && val != "closed"{
//		p.mu.RLock()
//		p.portScanRes[key]="filter"
//		p.mu.RUnlock()
//	}else{
//		p.mu.RLock()
//		p.portScanRes[key]="filter"
//		p.mu.RUnlock()
//	}
//}

// send sends the given layers as a single packet on the network.
func (p *tcpMsg) send(l ...gopacket.SerializableLayer) error {
	buffer := gopacket.NewSerializeBuffer()
	options := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}
	if err := gopacket.SerializeLayers(buffer, options, l...); err != nil {
		return err
	}
	return p.handle.WritePacketData(buffer.Bytes())
}

func (p *tcpMsg) Test(){
	jlog.Info("portScanRes")
	p.portScanRes.Range(func(key, value interface{}) bool {
		jlog.Info("key:",key,"val:",value)
		return true
	})
	jlog.Info("portScanTasks")
	//p.portScanTasks.Range(func(key, value interface{}) bool {
	//	jlog.Info("key:",key,"val:",value)
	//	return true
	//})
}

func (p *tcpMsg) RecvScanRes()(){
	for{
		//if status,ok := p.portScanRes.Load(key); ok{
		//	p.portScanTasks.Delete(key)
		//	p.portScanRes.Delete(key)
		//	return fmt.Sprintf("%v",remotePort),status.(string),nil
		//}
		//if time.Since(start) > p.timeout {
		//	//jlog.Error("start:",start.String(),remotePort,"超时",time.Since(start).String())
		//	return fmt.Sprintf("%v",remotePort),"filter",fmt.Errorf("timeout")
		//}
		// 读取数据包
		data, _, err := p.handle.ReadPacketData()
		if err == pcap.NextErrorTimeoutExpired {
			jlog.Error("readPacketData:1err:",err)
			continue
		} else if err != nil {
			jlog.Error("readPacketData:err:",err)
			continue
		}
		packet := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.NoCopy)

		// 网络层
		if jnet := packet.NetworkLayer(); jnet == nil {
		}else if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer == nil { // 传输层
			// log.Printf("packet has not tcp layer")
		}else if tcp, ok := tcpLayer.(*layers.TCP); !ok { // 解码成标准传输层
		}else if _,ok := p.portScanTasks.Load(jnet.NetworkFlow().Dst().String()+":"+tcp.DstPort.String()+"-"+jnet.NetworkFlow().Src().String()+":"+ strings.Split(tcp.SrcPort.String(),"(")[0] ); !ok{
			// 接收到的数据包的flow与已发送的flow不匹配
		}else  if tcp.RST {
			// 端口关闭
			p.portScanRes.Store(jnet.NetworkFlow().Dst().String()+":"+tcp.DstPort.String()+"-"+jnet.NetworkFlow().Src().String()+":"+strings.Split(tcp.SrcPort.String(),"(")[0],"closed")
			p.portScanTasks.Delete(jnet.NetworkFlow().Dst().String()+":"+tcp.DstPort.String()+"-"+jnet.NetworkFlow().Src().String()+":"+strings.Split(tcp.SrcPort.String(),"(")[0])
		} else if tcp.SYN && tcp.ACK {
			// 端口开放
			p.portScanRes.Store(jnet.NetworkFlow().Dst().String()+":"+tcp.DstPort.String()+"-"+jnet.NetworkFlow().Src().String()+":"+strings.Split(tcp.SrcPort.String(),"(")[0],"open")
			p.portScanTasks.Delete(jnet.NetworkFlow().Dst().String()+":"+tcp.DstPort.String()+"-"+jnet.NetworkFlow().Src().String()+":"+strings.Split(tcp.SrcPort.String(),"(")[0])
		}else{
			// 无效包
			jlog.Debug("xxxx")
			p.portScanTasks.Delete(jnet.NetworkFlow().Dst().String()+":"+tcp.DstPort.String()+"-"+jnet.NetworkFlow().Src().String()+":"+strings.Split(tcp.SrcPort.String(),"(")[0])
		}

		//if jnet := packet.NetworkLayer(); jnet == nil {
		//} else if jnet.NetworkFlow().String() != ipFlow.String() {
		//	// log.Printf("packet does not match our ip src/dst")
		//} else if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer == nil {
		//	// log.Printf("packet has not tcp layer")
		//} else if tcp, ok := tcpLayer.(*layers.TCP); !ok {
		//	//jlog.Error("tcp layer is not tcp layer :-/")
		//	return fmt.Sprintf("%v",remotePort),"",fmt.Errorf("tcp layer is not tcp layer")
		//	//} else if _,ok := p.portScanTasks[jnet.NetworkFlow().Dst().String()+":"+tcp.DstPort.String()+"-"+jnet.NetworkFlow().Src().String()+":"+tcp.SrcPort.String()]; !ok{
		//} else if _,ok := p.portScanTasks.Load(jnet.NetworkFlow().Dst().String()+":"+tcp.DstPort.String()+"-"+jnet.NetworkFlow().Src().String()+":"+ strings.Split(tcp.SrcPort.String(),"(")[0] ); !ok{
		//	// 接收到的数据包的flow与已发送的flow不匹配
		//}else  if tcp.RST {
		//	// 端口关闭
		//	p.portScanRes.Store(jnet.NetworkFlow().Dst().String()+":"+tcp.DstPort.String()+"-"+jnet.NetworkFlow().Src().String()+":"+strings.Split(tcp.SrcPort.String(),"(")[0],"closed")
		//	p.portScanTasks.Delete(jnet.NetworkFlow().Dst().String()+":"+tcp.DstPort.String()+"-"+jnet.NetworkFlow().Src().String()+":"+strings.Split(tcp.SrcPort.String(),"(")[0])
		//} else if tcp.SYN && tcp.ACK {
		//	// 端口开放
		//	p.portScanRes.Store(jnet.NetworkFlow().Dst().String()+":"+tcp.DstPort.String()+"-"+jnet.NetworkFlow().Src().String()+":"+strings.Split(tcp.SrcPort.String(),"(")[0],"open")
		//	p.portScanTasks.Delete(jnet.NetworkFlow().Dst().String()+":"+tcp.DstPort.String()+"-"+jnet.NetworkFlow().Src().String()+":"+strings.Split(tcp.SrcPort.String(),"(")[0])
		//}else{
		//	// 无效包
		//	jlog.Debug("xxxx")
		//	p.portScanTasks.Delete(jnet.NetworkFlow().Dst().String()+":"+tcp.DstPort.String()+"-"+jnet.NetworkFlow().Src().String()+":"+strings.Split(tcp.SrcPort.String(),"(")[0])
		//}

	}
}

// 发送tcp syn数据包
func (p *tcpMsg) SinglePortSYNScan(remoteIP string,remotePort uint16,payload string) (port string,status string,err error){
	// 数据链路层
	_srcMAC,err := net.ParseMAC(p.localNetworkInst.LocalMAC)
	if err != nil{
		jlog.Error(err)
		return "", "", err
	}
	_dstMAC,err := net.ParseMAC(p.remoteMAC)
	if err != nil{
		jlog.Error(err)
		return "","",err
	}
	ethernetLayer := &layers.Ethernet{
		SrcMAC: _srcMAC,
		DstMAC: _dstMAC, //
		EthernetType: layers.EthernetTypeIPv4,
	}
	// 网络层
	ipLayer := &layers.IPv4{
		SrcIP: net.ParseIP(p.localNetworkInst.LocalIP),
		DstIP: net.ParseIP(remoteIP),
		Version: 4,
		TTL:64,
		Protocol: layers.IPProtocolTCP,
	}
	// 传输层
	// 获取一个空闲的端口
	_srcPort,err := jcore.GetFreePort(p.localNetworkInst.LocalIP)
	if err != nil{
		for {
			_srcPort,err = jcore.GetFreePort(p.localNetworkInst.LocalIP)
			if err == nil{
				break
			}
		}
	}
	tcpLayer := &layers.TCP{
		SrcPort: layers.TCPPort(_srcPort),
		DstPort: layers.TCPPort(remotePort),
		SYN:true,
		//Window: 29200,
		Window: 1024,
		Options: []layers.TCPOption{
			layers.TCPOption{layers.TCPOptionKindMSS, 4,[]byte("\x05\xb4")},
			//layers.TCPOption{layers.TCPOptionKindSACKPermitted, 2, nil},
			//layers.TCPOption{layers.TCPOptionKindNop, 1, nil},
			//layers.TCPOption{layers.TCPOptionKindWindowScale, 3, []byte("\x07")},
		},
	}
	tcpLayer.SetNetworkLayerForChecksum(ipLayer)

	//err = p.send(ethernetLayer,ipLayer,tcpLayer,gopacket.Payload([]byte(payload)))
	err = p.send(ethernetLayer,ipLayer,tcpLayer)
	if err != nil{
		jlog.Error(err)
		return "","",err
	}
	key := p.localNetworkInst.LocalIP+":"+strconv.Itoa(int(_srcPort))+"-"+remoteIP+":"+strconv.Itoa(int(remotePort))

	p.portScanTasks.Store(key,1)
	start := time.Now()
	ipFlow := gopacket.NewFlow(layers.EndpointIPv4, net.ParseIP(remoteIP), net.ParseIP(p.localNetworkInst.LocalIP))

	for{
		if status,ok := p.portScanRes.Load(key); ok{
			p.portScanTasks.Delete(key)
			p.portScanRes.Delete(key)
			return fmt.Sprintf("%v",remotePort),status.(string),nil
		}
		if time.Since(start) > p.timeout {
			jlog.Warn(remotePort,"start:",start.String(),"超时",time.Since(start).String(),"p.timeout:",p.timeout)
			return fmt.Sprintf("%v",remotePort),"filter",fmt.Errorf("timeout")
		}
		//jlog.Error(remotePort,"start:",start.String(),"准备",time.Since(start).String(),"p.timeout:",p.timeout)
		data, _, err := p.handle.ReadPacketData()
		//jlog.Error(remotePort,"start:",start.String(),"获取到",time.Since(start).String(),"p.timeout:",p.timeout)
		if err == pcap.NextErrorTimeoutExpired {
			jlog.Error("readPacketData:1err:",remotePort,err)
			continue
		} else if err != nil {
			jlog.Error("readPacketData:err:",err)
			continue
		}
		packet := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.NoCopy)
		if jnet := packet.NetworkLayer(); jnet == nil {
		} else if jnet.NetworkFlow().String() != ipFlow.String() {
			// log.Printf("packet does not match our ip src/dst")
		} else if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer == nil {
			// log.Printf("packet has not tcp layer")
		} else if tcp, ok := tcpLayer.(*layers.TCP); !ok {
			//jlog.Error("tcp layer is not tcp layer :-/")
			return fmt.Sprintf("%v",remotePort),"",fmt.Errorf("tcp layer is not tcp layer")
		//} else if _,ok := p.portScanTasks[jnet.NetworkFlow().Dst().String()+":"+tcp.DstPort.String()+"-"+jnet.NetworkFlow().Src().String()+":"+tcp.SrcPort.String()]; !ok{
		} else if _,ok := p.portScanTasks.Load(jnet.NetworkFlow().Dst().String()+":"+tcp.DstPort.String()+"-"+jnet.NetworkFlow().Src().String()+":"+ strings.Split(tcp.SrcPort.String(),"(")[0] ); !ok{
			// 接收到的数据包的flow与已发送的flow不匹配
		}else  if tcp.RST {
			// 端口关闭
			p.portScanRes.Store(jnet.NetworkFlow().Dst().String()+":"+tcp.DstPort.String()+"-"+jnet.NetworkFlow().Src().String()+":"+strings.Split(tcp.SrcPort.String(),"(")[0],"closed")
			p.portScanTasks.Delete(jnet.NetworkFlow().Dst().String()+":"+tcp.DstPort.String()+"-"+jnet.NetworkFlow().Src().String()+":"+strings.Split(tcp.SrcPort.String(),"(")[0])
		} else if tcp.SYN && tcp.ACK {
			jlog.Debug("端口开放，",remotePort)
			// 端口开放
			p.portScanRes.Store(jnet.NetworkFlow().Dst().String()+":"+tcp.DstPort.String()+"-"+jnet.NetworkFlow().Src().String()+":"+strings.Split(tcp.SrcPort.String(),"(")[0],"open")
			p.portScanTasks.Delete(jnet.NetworkFlow().Dst().String()+":"+tcp.DstPort.String()+"-"+jnet.NetworkFlow().Src().String()+":"+strings.Split(tcp.SrcPort.String(),"(")[0])
		}else{
			// 无效包
			jlog.Debug("xxxx")
			p.portScanTasks.Delete(jnet.NetworkFlow().Dst().String()+":"+tcp.DstPort.String()+"-"+jnet.NetworkFlow().Src().String()+":"+strings.Split(tcp.SrcPort.String(),"(")[0])
		}

	}

}

func (p *tcpMsg)GetHWAddr(dstIPStr string) (GWMACStr string,err error){
	start := time.Now()
	// 数据链路层
	// Prepare the layers to send for an ARP request.
	_srcMAC,_ := net.ParseMAC(p.localNetworkInst.LocalMAC)
	eth := &layers.Ethernet{
		SrcMAC:       _srcMAC,
		DstMAC:       net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		EthernetType: layers.EthernetTypeARP,
	}
	_srcIP := net.ParseIP(p.localNetworkInst.LocalIP)
	_dstIP := net.ParseIP(dstIPStr)
	arp := &layers.ARP{
		AddrType:          layers.LinkTypeEthernet,
		Protocol:          layers.EthernetTypeIPv4,
		HwAddressSize:     6,
		ProtAddressSize:   4,
		Operation:         layers.ARPRequest,
		SourceHwAddress:   []byte(_srcMAC),
		SourceProtAddress: []byte(_srcIP.To4()),
		DstHwAddress:      []byte{0, 0, 0, 0, 0, 0},
		DstProtAddress:    []byte(_dstIP.To4()),
	}
	jlog.Debug("GateWayIP:",dstIPStr)
	if err := gopacket.SerializeLayers(p.buffer, p.options, eth,arp); err != nil {
		jlog.Fatal(err)
	}
	p.handle.WritePacketData(p.buffer.Bytes())

	// Wait 3 seconds for an ARP reply.
	for {
		if time.Since(start) > p.timeout {
			jlog.Error("timeout getting ARP reply")
			return "",fmt.Errorf(dstIPStr,"timeout getting getway mac")
		}
		data, _, err := p.handle.ReadPacketData()
		if err == pcap.NextErrorTimeoutExpired {
			continue
		} else if err != nil {
			continue
		}
		packet := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.NoCopy)
		if arpLayer := packet.Layer(layers.LayerTypeARP); arpLayer != nil {
			arp := arpLayer.(*layers.ARP)
			//jlog.Println(net.IP(arp.SourceProtAddress))
			//jlog.Println(net.ParseIP(dstIPStr))
			if net.IP(arp.SourceProtAddress).Equal(net.ParseIP(dstIPStr)) {
				//return net.HardwareAddr(arp.SourceHwAddress), nil
				jlog.Debug(net.HardwareAddr(arp.SourceHwAddress))
				GWMACStr = net.HardwareAddr(arp.SourceHwAddress).String()
				return GWMACStr,nil
			}
		}
	}
}
