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
	p.handle, err = pcap.OpenLive(p.localNetworkInst.LocalDevice, p.snapshot_len, p.promiscuous, p.timeout)
	if err != nil {jlog.Fatal(err) }
}

func (p *tcpMsg)CloseHandle(){
	p.handle.Close()
}

func (p *tcpMsg) SinglePortSYNScan(remoteIP string,remotePort uint16,payload string) (port string,status string,err error){
	handle2, err := pcap.OpenLive(p.localNetworkInst.LocalDevice, p.snapshot_len, p.promiscuous, p.timeout)
	if err != nil {jlog.Fatal(err) }
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

	buffer := gopacket.NewSerializeBuffer()
	options := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}
	// And create the packet with the layers
	err = gopacket.SerializeLayers(buffer, options,
		ethernetLayer,
		ipLayer,
		tcpLayer,
		gopacket.Payload([]byte(payload)),
	)
	if err != nil{
		jlog.Error(err)
		return "","",err
	}
	outgoingPacket := buffer.Bytes()
	//err = p.handle.WritePacketData(outgoingPacket)
	err = handle2.WritePacketData(outgoingPacket)
	if err != nil{
		jlog.Error(err)
		return "","",err
	}
	start := time.Now()
	ipFlow := gopacket.NewFlow(layers.EndpointIPv4, net.ParseIP(remoteIP), net.ParseIP(p.localNetworkInst.LocalIP))
	for{
		if time.Since(start) > p.timeout {
			//jlog.Printf("port %v filter\n", remotePort)
			handle2.Close()
			return fmt.Sprintf("%v",remotePort),"filter",fmt.Errorf("timeout")
		}
		//data, _, err := p.handle.ReadPacketData()
		data, _, err := handle2.ReadPacketData()
		//data,_,err := p.handle.ZeroCopyReadPacketData()
		if err == pcap.NextErrorTimeoutExpired {
			continue
		} else if err != nil {
			//jlog.Error("error reading packet: %v", err)
			continue
		}
		//packet := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.Default)
		packet := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.NoCopy)
		if jnet := packet.NetworkLayer(); jnet == nil {
		} else if jnet.NetworkFlow().String() != ipFlow.String() {
			// log.Printf("packet does not match our ip src/dst")
		} else if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer == nil {
			// log.Printf("packet has not tcp layer")
		} else if tcp, ok := tcpLayer.(*layers.TCP); !ok {
			// We panic here because this is guaranteed to never
			// happen.
			//jlog.Error("tcp layer is not tcp layer :-/")
			return fmt.Sprintf("%v",remotePort),"",fmt.Errorf("tcp layer is not tcp layer")
		} else if tcp.DstPort != layers.TCPPort(_srcPort) {
			// log.Printf("dst port %v does not match", tcp.DstPort)
		} else if tcp.RST {
			//jlog.Printf("port %v closed\n", tcp.SrcPort)
			return tcp.SrcPort.String(),"closed",nil
		} else if tcp.SYN && tcp.ACK {
			//jlog.Printf("port %v open\n", tcp.SrcPort)
			return tcp.SrcPort.String(),"open",nil
		} else {
			// log.Printf("ignoring useless packet")
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
