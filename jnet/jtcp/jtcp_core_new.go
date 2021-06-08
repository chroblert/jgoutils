// +build ignore

package jtcp

import (
	"fmt"
	"github.com/chroblert/jgoutils/jlog"
	"github.com/chroblert/jgoutils/jnet/jtcp/jcore"
	"github.com/chroblert/jgoutils/jnet/jtcp/jroute"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/pfring"
	"github.com/petermattis/goid"
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

	handle       *pfring.Ring
	buffer       gopacket.SerializeBuffer
	options      gopacket.SerializeOptions

	remoteMAC string
	//portScanRes map[string]string // localIP:localPort-remoteIP:remotePort,status
	portScanRes *sync.Map // localIP:localPort-remoteIP:remotePort,status
	mu *sync.RWMutex
	//portScanTasks map[string]
	portScanTasks *sync.Map
	portScanCount int
	isRecvScanRes bool

}

func New() *tcpMsg {
	tmp := jcore.GetNetWorks()
	tm := &tcpMsg{
		localNetworkInst: tmp[0],
		snapshot_len:     1024,
		promiscuous:      false,
		timeout:          500*time.Millisecond,
		handle: &pfring.Ring{},
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
		portScanCount: 0,
		isRecvScanRes:false,
	}
	for i := 0; i < len(tmp);i++{
		if err := tm.SetNetwork(i); err == nil{
			break
		}else if i == len(tmp) -1{
			jlog.Error("获取网关的MAC地址失败")
			return nil
		}
	}
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
	//p.handle, err = pcap.OpenLive(p.localNetworkInst.LocalDevice, p.snapshot_len, p.promiscuous, p.timeout)
	p.handle,err = pfring.NewRing(p.localNetworkInst.LocalDevice,uint32(p.snapshot_len),pfring.FlagPromisc)
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
		if value == "open"{
			jlog.Info("key:",key,"val:",value)
		}
		return true
	})
	jlog.Info("portScanTasks")
	//p.portScanTasks.Range(func(key, value interface{}) bool {
	//	jlog.Info("key:",key,"val:",value)
	//	return true
	//})
}

//func (p *tcpMsg) RecvScanRes(start time.Time,ipStr string,remotePort string,localPort string)(portStr string,status string,err error){
func (p *tcpMsg) RecvScanRes(remoteIP string,remotePort string,localPort string){
	handle, err := pcap.OpenLive(p.localNetworkInst.LocalDevice, p.snapshot_len, p.promiscuous, p.timeout)
	if err != nil{
		jlog.Error(err)
	}
	defer func() {
		jlog.Warn("退出")
		handle.Close()
	}()
	jlog.Debug("接收端口扫描结果")
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	var eth layers.Ethernet
	var ipv4 layers.IPv4
	var ipv6 layers.IPv6
	var tcp layers.TCP
	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet, &eth, &ipv4, &ipv6, &tcp)

outloop:
	for{
		packet, err := packetSource.NextPacket()
		if err != nil{
			continue
		}
		var decoded []gopacket.LayerType
		err = parser.DecodeLayers(packet.Data(), &decoded)
		if err != nil{
			continue
		}
		isContainTcp := false
		for _,v := range decoded{
			if v == layers.LayerTypeTCP{
				isContainTcp = true
				break
			}
		}
		if !isContainTcp {
			continue outloop
		}
		if ipv4.SrcIP.String() == p.localNetworkInst.LocalIP {
			continue outloop
		}
		key := ipv4.DstIP.String()+":"+strings.Split(tcp.DstPort.String(),"(")[0]+"-"+ipv4.SrcIP.String()+":"+strings.Split(tcp.SrcPort.String(),"(")[0]
		//if tcp.SrcPort == layers.TCPPort(22){
		//	jlog.Warn(tcp)
		//}
		if _,ok := p.portScanTasks.LoadAndDelete(key);ok{
			if tcp.ACK && tcp.SYN{
				jlog.Warn(goid.Get(),":",ipv4.SrcIP,":",tcp.SrcPort,"open",string(tcp.Payload))
			}
		}

	}
	//for{
	//	//p.mu.RLock()
	//	//tmpPortScanCount := p.portScanCount
	//	//p.mu.RUnlock()
	//	////jlog.Warn(goid.Get(),":",tmpPortScanCount)
	//	//if tmpPortScanCount < 1{
	//	//	jlog.Warn(goid.Get(),"接收完成")
	//	//	return
	//	//}
	//	//if time.Since(start) > 3*time.Second{
	//	//	return
	//	//}
	//	// 读取数据包
	//	data, _, err := handle.ReadPacketData()
	//	if err == pcap.NextErrorTimeoutExpired {
	//		continue
	//	} else if err != nil {
	//		continue
	//	}
	//	packet := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.NoCopy)
	//
	//	// 网络层
	//	if jnet := packet.NetworkLayer(); jnet == nil {
	//		continue
	//	}else if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer == nil { // 传输层
	//		// log.Printf("packet has not tcp layer")
	//		continue
	//	}else if tcp, ok := tcpLayer.(*layers.TCP); !ok { // 解码成标准传输层
	//		continue
	//	}else if _,ok := p.portScanTasks.Load(jnet.NetworkFlow().Dst().String()+":"+tcp.DstPort.String()+"-"+jnet.NetworkFlow().Src().String()+":"+ strings.Split(tcp.SrcPort.String(),"(")[0] ); !ok{
	//		// 接收到的数据包的flow与已发送的flow不匹配
	//		continue
	//	}else  if tcp.RST {
	//		//jlog.Warn(goid.Get(),":",strings.Split(tcp.SrcPort.String(),"(")[0]+":"+" closed")
	//		// 端口关闭
	//		p.portScanRes.Store(jnet.NetworkFlow().Dst().String()+":"+tcp.DstPort.String()+"-"+jnet.NetworkFlow().Src().String()+":"+strings.Split(tcp.SrcPort.String(),"(")[0],"closed")
	//		p.portScanTasks.Delete(jnet.NetworkFlow().Dst().String()+":"+tcp.DstPort.String()+"-"+jnet.NetworkFlow().Src().String()+":"+strings.Split(tcp.SrcPort.String(),"(")[0])
	//	} else if tcp.SYN && tcp.ACK {
	//		jlog.Warn(goid.Get(),":",jnet.NetworkFlow().Src().String(),strings.Split(tcp.SrcPort.String(),"(")[0]+":"+" open")
	//		//jlog.Warn(goid.Get(),":",jnet.NetworkFlow().Dst().String()+":"+tcp.DstPort.String()+"-"+jnet.NetworkFlow().Src().String()+":"+ strings.Split(tcp.SrcPort.String(),"(")[0] )
	//		// 端口开放
	//		p.portScanRes.Store(jnet.NetworkFlow().Dst().String()+":"+tcp.DstPort.String()+"-"+jnet.NetworkFlow().Src().String()+":"+strings.Split(tcp.SrcPort.String(),"(")[0],"open")
	//		p.portScanTasks.Delete(jnet.NetworkFlow().Dst().String()+":"+tcp.DstPort.String()+"-"+jnet.NetworkFlow().Src().String()+":"+strings.Split(tcp.SrcPort.String(),"(")[0])
	//	}else{
	//		// 无效包
	//		jlog.Debug(goid.Get(),":","xxxx")
	//		p.portScanTasks.Delete(jnet.NetworkFlow().Dst().String()+":"+tcp.DstPort.String()+"-"+jnet.NetworkFlow().Src().String()+":"+strings.Split(tcp.SrcPort.String(),"(")[0])
	//	}
	//}
}

// 发送tcp syn数据包
func (p *tcpMsg) SinglePortSYNScan(remoteIP string,remotePort uint16) (ipStr,port string,status string,err error){
	// 数据链路层
	_srcMAC,err := net.ParseMAC(p.localNetworkInst.LocalMAC)
	if err != nil{
		jlog.Error(err)
		return "","", "", err
	}
	_dstMAC,err := net.ParseMAC(p.remoteMAC)
	if err != nil{
		jlog.Error(err)
		return "","","",err
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
			layers.TCPOption{layers.TCPOptionKindSACKPermitted, 2, nil},
			layers.TCPOption{layers.TCPOptionKindNop, 1, nil},
			layers.TCPOption{layers.TCPOptionKindWindowScale, 3, []byte("\x07")},
		},
	}
	tcpLayer.SetNetworkLayerForChecksum(ipLayer)
	key := p.localNetworkInst.LocalIP+":"+strconv.Itoa(int(_srcPort))+"-"+remoteIP+":"+strconv.Itoa(int(remotePort))
	p.portScanTasks.Store(key,1)
	//err = p.send(ethernetLayer,ipLayer,tcpLayer,gopacket.Payload([]byte(payload)))
	err = p.send(ethernetLayer,ipLayer,tcpLayer)
	if err != nil{
		jlog.Error(err)
		return "","","",err
	}
	time.Sleep(1500*time.Millisecond)

	//jlog.Warn(goid.Get(),": send :",remotePort)
	//p.mu.Lock()
	//p.portScanCount ++
	//p.mu.Unlock()
	//p.mu.RLock()
	//if p.portScanCount == 1{
	//	jlog.Warn("go p.RecvScanRes()")

	//var wg sync.WaitGroup
	//wg.Add(1)
	//go func(remoteIP string,remotePort string,localPort string) {
	//	p.RecvScanRes(remoteIP ,remotePort ,localPort )
	//	wg.Done()
	//}(remoteIP,strconv.Itoa(int(remotePort)),strconv.Itoa(int(_srcPort)))
	//wg.Wait()
	//}
	//p.mu.RUnlock()




	//start := time.Now()
	//for{
	//	if 	val,ok := p.portScanRes.Load(key);ok{
	//		//if val == "open"{
	//		//	jlog.Debug(remotePort,":",val)
	//		//}
	//		//p.portScanTasks.Delete(key)
	//		return strconv.Itoa(int(remotePort)), val.(string), nil
	//	}
	//	if time.Since(start) > 1*time.Second {
	//		//jlog.Debug(time.Since(start))
	//		//p.portScanTasks.Delete(key)
	//		return strconv.Itoa(int(remotePort)), "filter", fmt.Errorf("timeout")
	//	}
	//}
	//start := time.Now()
	//defer func() {
	//	jlog.Warn(goid.Get(),":退出")
	//}()
	//for{
	//	//p.portScanRes.Range(func(key1, value interface{}) bool {
	//	//	if key == key1{
	//	//		jlog.Warn(goid.Get()," : ",key,":",value)
	//	//		p.mu.Lock()
	//	//		p.portScanCount--
	//	//		p.mu.Unlock()
	//	//		return false
	//	//	}
	//	//	return true
	//	//})
	//	if val,ok := p.portScanRes.Load(key);ok {
	//		jlog.Warn(goid.Get(),":",remotePort,val)
	//		p.mu.Lock()
	//		p.portScanCount--
	//		p.mu.Unlock()
	//		return strconv.Itoa(int(remotePort)), val.(string), nil
	//	}
	//	if time.Since(start) > 1*time.Second {
	//		//jlog.Debug(time.Since(start))
	//		//p.portScanTasks.Delete(key)
	//		p.mu.Lock()
	//		p.portScanCount--
	//		p.mu.Unlock()
	//		//jlog.Warn(goid.Get(),":超时，",remotePort)
	//		return strconv.Itoa(int(remotePort)), "filter", fmt.Errorf("timeout")
	//	}
	//}

	return "","", "", nil
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
	//if err := gopacket.SerializeLayers(p.buffer, p.options, eth,arp); err != nil {
	//	jlog.Fatal(err)
	//}
	//p.handle.WritePacketData(p.buffer.Bytes())
	p.send(eth,arp)

	// Wait 3 seconds for an ARP reply.
	for {
		if time.Since(start) > p.timeout*2 {
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
			//jlog.Debug("src:",arp.SourceProtAddress)
			//jlog.Debug("dst:",arp.DstProtAddress)
			if net.IP(arp.SourceProtAddress).Equal(net.ParseIP(dstIPStr)) {
				//jlog.Debug(net.HardwareAddr(arp.SourceHwAddress))
				GWMACStr = net.HardwareAddr(arp.SourceHwAddress).String()
				return GWMACStr,nil
			}
		}
	}
}
