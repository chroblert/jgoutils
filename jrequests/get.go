package jrequests

func Get(requrl string,opts...OptionInterface) (statuscode int,headers map[string][]string,body []byte,err error){
	getOption := getDefaultOptions()

	for _,opt := range opts{
		opt.apply(&getOption)
	}
	//fmt.Println(getOption)
	statuscode,headers,body,err = SingleReq("GET",requrl,getOption)
	return statuscode,headers,body,err
}

