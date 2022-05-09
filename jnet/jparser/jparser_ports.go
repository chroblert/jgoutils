package jparser

import (
	"fmt"
	"strings"
)

// port1,port2-port3
func ParsePortStr(portStr string) []int {
	tmpList := strings.Split(portStr, ",")
	portList := make([]int, 0)
	for _, v := range tmpList {
		if ports, err := getPortsFromPortRange(v); err == nil {
			portList = append(portList, ports...)
		} else if 1 <= string2Int(v) && string2Int(v) <= 65535 {
			portList = append(portList, string2Int(v))
		} else {
		}
	}
	tmpList = tmpList[:0]
	return portList
}

func getPortsFromPortRange(portRangeStr string) ([]int, error) {
	tmpList := strings.Split(portRangeStr, "-")
	if len(tmpList) != 2 {
		return nil, fmt.Errorf("不符合格式。port1-port2")
	}
	startPort := string2Int(tmpList[0])
	endPort := string2Int(tmpList[1])
	if startPort > endPort {
		return nil, fmt.Errorf("port1应小于port2")
	}
	if 65535 < startPort || startPort < 1 || 65535 < endPort || endPort < 1 {
		return nil, fmt.Errorf("port 应在1-65535之间")
	}
	portList := make([]int, 0)
	for port := startPort; port <= endPort; port++ {
		portList = append(portList, port)
	}
	tmpList = tmpList[:0]
	return portList, nil
}
