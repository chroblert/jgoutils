package jparser

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// ip1,ip2,ip3-ip4,ip5/cidr
func ParseIPStr(ipStr string) []string {
	tmpList := strings.Split(ipStr, ",")
	ipStrList := make([]string, 0)
	for _, v := range tmpList {
		// 是不是CIDR
		if ips, err := getIPSFromCIDR(v); err == nil {
			ipStrList = append(ipStrList, ips...)
		} else if ips, err := getIPSFromIPRange(v); err == nil {
			ipStrList = append(ipStrList, ips...)
		} else if ip := net.ParseIP(v); ip != nil {
			ipStrList = append(ipStrList, ip.String())
		} else {
			//jlog.Error(err)
		}
	}
	// 清空切片
	tmpList = tmpList[:0]
	//sort.Strings(ipStrList)
	return removeDuplicateElement(ipStrList)
}

func removeDuplicateElement(addrs []string) []string {
	result := make([]string, 0, len(addrs))
	temp := map[string]struct{}{}
	for _, item := range addrs {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

func getIPSFromIPRange(ipRange string) ([]string, error) {
	tmpList := strings.Split(ipRange, "-")
	if len(tmpList) != 2 {
		return nil, fmt.Errorf("不符合格式。ip1-ip2")
	}
	if ip2Int(tmpList[0]) > ip2Int(tmpList[1]) {
		return nil, fmt.Errorf("ip1应小于ip2")
	}
	ips := make([]string, 0)
	startIP := net.ParseIP(tmpList[0])
	endIP := net.ParseIP(tmpList[1])
	// 过滤最后一位为0和255的IP
	for ip := startIP; ip2Int(ip.String()) <= ip2Int(endIP.String()); inc(ip) {
		if ip.To4()[3] == 255 || ip.To4()[3] == 0 {
			continue
		}
		ips = append(ips, ip.String())
	}
	return ips, nil
}

func ip2Int(ip string) int64 {
	if len(ip) == 0 {
		return 0
	}
	bits := strings.Split(ip, ".")
	if len(bits) < 4 {
		return 0
	}
	b0 := string2Int(bits[0])
	b1 := string2Int(bits[1])
	b2 := string2Int(bits[2])
	b3 := string2Int(bits[3])

	var sum int64
	sum += int64(b0) << 24
	sum += int64(b1) << 16
	sum += int64(b2) << 8
	sum += int64(b3)

	return sum
}

func string2Int(in string) (out int) {
	out, _ = strconv.Atoi(in)
	return
}

func getIPSFromCIDR(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}
	if len(ips) > 1 {
		return ips[1 : len(ips)-1], nil
	} else {
		return nil, fmt.Errorf("no element")
	}
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
