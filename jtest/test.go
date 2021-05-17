package jtest

import (
	"github.com/chroblert/JC-GoUtils/jconv"
	"github.com/chroblert/JC-GoUtils/jlog"
)

func init() {
	//tmp := [][]interface{}{{"1t1","1t2","1t3"},{"2t1","2t2","2t3","2t4"},{"3t1","3t2"}}
	tmp := [][]string{{"1t1", "1t2", "1t3"}, {"2t1", "2t2", "2t3", "2t4"}, {"3t1", "3t2"}}
	ttt := Dikaer(jconv.ConvertStringssToInterfacess(tmp))
	for _, v := range ttt {
		jlog.Info(v)
	}
}

// 获取多个数组的排列组合，即笛卡尔积
func Dikaer(sets [][]interface{}) [][]interface{} {
	lens := func(i int) int { return len(sets[i]) }
	product := [][]interface{}{}
	for ix := make([]int, len(sets)); ix[0] < lens(0); nextIndex(ix, lens) {
		var r []interface{}
		for j, k := range ix {
			r = append(r, sets[j][k])
		}
		product = append(product, r)
	}
	return product
}

func nextIndex(ix []int, lens func(i int) int) {
	for j := len(ix) - 1; j >= 0; j-- {
		ix[j]++
		if j == 0 || ix[j] < lens(j) {
			return
		}
		ix[j] = 0
	}
}
