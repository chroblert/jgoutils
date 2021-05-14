package jlog


func (buf *buffer) write2(i, d int) {
	buf.temp[i+1] = digits[d%10]
	d /= 10
	buf.temp[i] = digits[d%10]
}

func (buf *buffer) write4(i, d int) {
	buf.temp[i+3] = digits[d%10]
	d /= 10
	buf.temp[i+2] = digits[d%10]
	d /= 10
	buf.temp[i+1] = digits[d%10]
	d /= 10
	buf.temp[i] = digits[d%10]
}

func (buf *buffer) writeN(i, d int) int {
	j := len(buf.temp)
	for d > 0 {
		j--
		buf.temp[j] = digits[d%10]
		d /= 10
	}
	return copy(buf.temp[i:], buf.temp[j:])
}
