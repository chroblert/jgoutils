package jrequests

//func Get(requrl string,opts...OptionInterface) (statuscode int,headers map[string][]string,body []byte,err error){
func Get(requrl string, opts ...OptionInterface) (statuscode int, respheaders map[string][]string, body []byte, err error) {
	getOption := getDefaultOptions()

	for _, opt := range opts {
		opt.apply(&getOption)
	}
	//fmt.Println(getOption)
	return SingleReq("GET", requrl, getOption)
	//return statuscode,headers,body,err
}

//func Post(requrl string,opts...OptionInterface) (statuscode int,headers map[string][]string,body []byte,err error){
func Post(requrl string, opts ...OptionInterface) (statuscode int, respheaders map[string][]string, body []byte, err error) {
	getOption := getDefaultOptions()
	for _, opt := range opts {
		opt.apply(&getOption)
	}
	//fmt.Println(getOption)
	return SingleReq("POST", requrl, getOption)
	//return statuscode,headers,body,err
}
