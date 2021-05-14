package jtest

import "fmt"

const (
	test = 1
)
var(
	test2 = 2
)
func init(){
	fmt.Println("test:",test)
	fmt.Println("test2:",test2)
	testfunc()
}

func testfunc(){
	fmt.Println("in testFunc")
}
func main(){
	fmt.Println("xxxx")
}