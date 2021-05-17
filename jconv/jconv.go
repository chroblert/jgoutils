package jconv

// 将字符串二维切片转换为接口二维切片
func ConvertStringssToInterfacess(stringss [][]string) (interfacess [][]interface{}) {
	//interfacess := [][]interface{}{}
	for _, v := range stringss {
		interfaces := []interface{}{}
		for _, v2 := range v {
			interfaces = append(interfaces, v2)
		}
		interfacess = append(interfacess, interfaces)
	}
	return
}
