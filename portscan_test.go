package main

import "testing"

func Benchmark_PortScan(b *testing.B){
	for i:=0;i<b.N;i++{
		portScan("101.132.112.169,1.15.178.39","1-650",100)
	}
}
