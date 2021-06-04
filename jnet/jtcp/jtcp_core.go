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
		timeout:          3 * time.Second,
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
	rt := jroute.NewRouteTable()
	gwIPStr,err := rt.GetGatewayByDstIP(tm.localNetworkInst.LocalIP)
	if err != nil{
		jlog.Error(err)
		return nil
	}
	tm.SetNetwork(0)
	gwMacStr,err := tm.GetHWAddr(gwIPStr)
	if err != nil{
		jlog.Error(err)
		for i := 1;i<len(tmp);i++{
			tm.SetNetwork(i)
			gwIPStr,err := rt.GetGatewayByDstIP(tm.localNetworkInst.LocalIP)
			if err != nil{
				jlog.Error(err)
				return nil
			}
			gwMacStr,err := tm.GetHWAddr(gwIPStr)
			if err == nil{
				tm.remoteMAC = gwMacStr
				break
			}
		}
	}
	tm.remoteMAC = gwMacStr
	return tm
}

func (p *tcpMsg)SetNetwork(id int){
	tmp := jcore.GetNetWorks()
	if len(tmp) > id{
		p.localNetworkInst = tmp[id]
	}
	if p.handle != nil{
		p.handle.Close()
	}
	var err error
	p.handle, err = pcap.OpenLive(p.localNetworkInst.LocalDevice, p.snapshot_len, p.promiscuous, pcap.BlockForever)
	if err != nil {jlog.Fatal(err) }
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
		Window: 29200,
		Options: []layers.TCPOption{
			layers.TCPOption{layers.TCPOptionKindMSS, 4,[]byte("\x05\xb4")},
			layers.TCPOption{layers.TCPOptionKindSACKPermitted, 2, nil},
			layers.TCPOption{layers.TCPOptionKindNop, 1, nil},
			layers.TCPOption{layers.TCPOptionKindWindowScale, 3, []byte("\x07")},
		},
	}
	tcpLayer.SetNetworkLayerForChecksum(ipLayer)

	err = p.send(ethernetLayer,ipLayer,tcpLayer,gopacket.Payload([]byte(payload)))
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
			return fmt.Sprintf("%v",remotePort),status.(string),nil
		}
		if time.Since(start) > p.timeout {
			return fmt.Sprintf("%v",remotePort),"filter",fmt.Errorf("timeout")
		}
		data, _, err := p.handle.ReadPacketData()
		if err == pcap.NextErrorTimeoutExpired {
			continue
		} else if err != nil {
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
			//jlog.Info("test")
			//jlog.Info(jnet.NetworkFlow().String())
			//jlog.Info(strings.Split(tcp.SrcPort.String(),"(")[0])
			//jlog.Info(jnet.NetworkFlow().Dst().String()+":"+tcp.DstPort.String()+"-"+jnet.NetworkFlow().Src().String()+":"+strings.Split(tcp.SrcPort.String(),"(")[0])
			//jlog.Info(p.portScanTasks.Load(jnet.NetworkFlow().Dst().String()+":"+tcp.DstPort.String()+"-"+jnet.NetworkFlow().Src().String()+":"+strings.Split(tcp.SrcPort.String(),"(")[0]))
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
	jlog.Debug(dstIPStr)
	if err := gopacket.SerializeLayers(p.buffer, p.options, eth,arp); err != nil {
		jlog.Fatal(err)
	}
	p.handle.WritePacketData(p.buffer.Bytes())

	// Wait 3 seconds for an ARP reply.
	for {
		if time.Since(start) > time.Second*4 {
			jlog.Error("timeout getting ARP reply")
			return "",fmt.Errorf("timeout getting ARP reply")
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
