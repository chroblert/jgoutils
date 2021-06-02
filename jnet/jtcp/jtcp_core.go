package jtcp

import (
	"fmt"
	"github.com/chroblert/jgoutils/jlog"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/routing"
	"log"
	"net"
	"runtime"
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
	localNetworkInst *localNetwork
	snapshot_len int32
	promiscuous bool
	timeout time.Duration

	handle       *pcap.Handle
	buffer       gopacket.SerializeBuffer
	options      gopacket.SerializeOptions


}

func New() *tcpMsg {
	tmp := GetNetWorks()
	return &tcpMsg{
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
}

func (p *tcpMsg)SetNetwork(id int){
	tmp := GetNetWorks()
	if len(tmp) > id{
		p.localNetworkInst = tmp[id]
	}
	if p.handle != nil{
		p.handle.Close()
	}
	var err error
	p.handle, err = pcap.OpenLive(p.localNetworkInst.localDevice, p.snapshot_len, p.promiscuous, p.timeout)
	if err != nil {log.Fatal(err) }
}

func (p *tcpMsg)CloseHandle(){
	p.handle.Close()
}

func (p *tcpMsg)SendSYN(remoteIP string,remotePort uint16,payload string) (port string,status string,err error){
	//var err error
	systype := runtime.GOOS
	if systype == "linux"{
		router,err := routing.New()
		if err != nil{
			jlog.Fatal(err)
		}
		iface, gw, src, err := router.Route(net.ParseIP(remoteIP))
		if err != nil {
			jlog.Fatal(err)
		}
		jlog.Println(iface,gw,src)
	}


	// 数据链路层
	_srcMAC,err := net.ParseMAC(p.localNetworkInst.localMAC)
	if err != nil{
		jlog.Fatal(err)
	}
	_dstMAC,err := net.ParseMAC("48:3f:e9:84:fc:c6")
	if err != nil{
		jlog.Fatal(err)
	}
	ethernetLayer := &layers.Ethernet{
		SrcMAC: _srcMAC,
		DstMAC: _dstMAC, //
		EthernetType: layers.EthernetTypeIPv4,
	}
	// 网络层
	ipLayer := &layers.IPv4{
		SrcIP: net.ParseIP(p.localNetworkInst.localIP),
		DstIP: net.ParseIP(remoteIP),
		Version: 4,
		TTL:64,
		Protocol: layers.IPProtocolTCP,
	}
	// 传输层
	_srcPort,_ := GetFreePort(p.localNetworkInst.localIP)
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

	// And create the packet with the layers
	err = gopacket.SerializeLayers(p.buffer, p.options,
		ethernetLayer,
		ipLayer,
		tcpLayer,
		gopacket.Payload([]byte(payload)),
	)
	if err != nil{
		jlog.Fatal(err)
	}
	outgoingPacket := p.buffer.Bytes()
	err = p.handle.WritePacketData(outgoingPacket)
	if err != nil{
		jlog.Fatal(err)
	}
	start := time.Now()
	ipFlow := gopacket.NewFlow(layers.EndpointIPv4, net.ParseIP(remoteIP), net.ParseIP(p.localNetworkInst.localIP))
	for{
		if time.Since(start) > p.timeout {
			jlog.Info("timeout getting  reply")
			return fmt.Sprintf("%v",remotePort),"",fmt.Errorf("timeout")
		}
		data, _, err := p.handle.ReadPacketData()
		if err == pcap.NextErrorTimeoutExpired {
			continue
		} else if err != nil {
			jlog.Error("error reading packet: %v", err)
			continue
		}
		packet := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.NoCopy)
		if jnet := packet.NetworkLayer(); jnet == nil {
		} else if jnet.NetworkFlow().String() != ipFlow.String() {
			// log.Printf("packet does not match our ip src/dst")
		} else if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer == nil {
			// log.Printf("packet has not tcp layer")
		} else if tcp, ok := tcpLayer.(*layers.TCP); !ok {
			// We panic here because this is guaranteed to never
			// happen.
			jlog.Error("tcp layer is not tcp layer :-/")
			return fmt.Sprintf("%v",remotePort),"",fmt.Errorf("tcp layer is not tcp layer")
		} else if tcp.DstPort != layers.TCPPort(_srcPort) {
			// log.Printf("dst port %v does not match", tcp.DstPort)
		} else if tcp.RST {
			jlog.Printf("  port %v closed", tcp.SrcPort)
			return tcp.SrcPort.String(),"closed",nil
		} else if tcp.SYN && tcp.ACK {
			jlog.Printf("  port %v open", tcp.SrcPort)
			return tcp.SrcPort.String(),"open",nil
		} else {
			// log.Printf("ignoring useless packet")
		}
	}
	return fmt.Sprintf("%v",remotePort),"",fmt.Errorf("timeout")

}

func (p *tcpMsg)GetHWAddr(dstIPStr string){
	//var err error
	//p.handle, err = pcap.OpenLive(p.localNetworkInst.localDevice, p.snapshot_len, p.promiscuous, p.timeout)
	//if err != nil {log.Fatal(err) }
	//defer p.handle.Close()
	start := time.Now()
	// 数据链路层
	// Prepare the layers to send for an ARP request.
	_srcMAC,_ := net.ParseMAC(p.localNetworkInst.localMAC)
	eth := &layers.Ethernet{
		SrcMAC:       _srcMAC,
		DstMAC:       net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		EthernetType: layers.EthernetTypeARP,
	}
	_srcIP := net.ParseIP(p.localNetworkInst.localIP)
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
		if time.Since(start) > time.Second*3 {
			jlog.Error("timeout getting ARP reply")
			return
		}
		data, _, err := p.handle.ReadPacketData()
		if err == pcap.NextErrorTimeoutExpired {
			continue
		} else if err != nil {
			return
		}
		packet := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.NoCopy)
		if arpLayer := packet.Layer(layers.LayerTypeARP); arpLayer != nil {
			arp := arpLayer.(*layers.ARP)
			jlog.Println(net.IP(arp.SourceProtAddress))
			jlog.Println(net.ParseIP(dstIPStr))
			if net.IP(arp.SourceProtAddress).Equal(net.ParseIP(dstIPStr)) {
				//return net.HardwareAddr(arp.SourceHwAddress), nil
				jlog.Debug(net.HardwareAddr(arp.SourceHwAddress))
			}
		}
	}
}
