package jrequests

// 请求选项的结构体
type Option struct {
	Proxy string
	Timeout int
	Headers map[string]string
	Data []byte
	Params map[string]string
	Cookies map[string]string
	IsRedirect bool
	IsVerifySSL bool
}

// 一个接口
type OptionInterface interface{
	apply(*Option)
}


type funcOption struct {
	f func(*Option)
}

func(fdo *funcOption) apply(option *Option){
	fdo.f(option)
}

//
func newFuncOption(f2 func(*Option)) *funcOption{
	return &funcOption{
		f:f2,
	}
}

// 设置代理
func SetProxy(s string) OptionInterface{
	return  newFuncOption(func(o *Option){
		o.Proxy = s
	})
}

// 设置超时
func SetTimeout(timeout int) OptionInterface{
	return newFuncOption(func(o *Option){
		o.Timeout = timeout
	})
}
// 设置headers
func SetHeaders(headers map[string]string) OptionInterface{
	return newFuncOption(func(o *Option){
		o.Headers = headers
	})
}
// 设置data
func SetData(data []byte) OptionInterface{
	return newFuncOption(func(o *Option){
		o.Data = data
	})
}
// 设置params
func SetParams(params map[string]string) OptionInterface{
	return newFuncOption(func(o *Option){
		o.Params = params
	})
}
// 设置cookies
func SetCookies(cookie map[string]string) OptionInterface{
	return newFuncOption(func(o *Option){
		o.Cookies = cookie
	})
}
// 设置是否转发
func SetIsRedirect(isredirect bool) OptionInterface{
	return newFuncOption(func(o *Option){
		o.IsRedirect = isredirect
	})
}

// 设置是否转发
func SetIsVerifySSL(isverifyssl bool) OptionInterface{
	return newFuncOption(func(o *Option){
		o.IsVerifySSL = isverifyssl
	})
}

// 获取默认设置
func getDefaultOptions() Option{
	return Option{
		Proxy: "",
		Timeout: 0,
		Headers: map[string]string{
			"User-Agent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:86.0) Gecko/20100101 Firefox/86.0",
		},
		Data: nil,
		Params: nil,
		Cookies: nil,
		IsRedirect: false,
		IsVerifySSL: false,
	}
}
