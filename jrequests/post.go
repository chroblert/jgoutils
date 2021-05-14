package jrequests

func Post(requrl string,opts...OptionInterface) (statuscode int,headers map[string][]string,body []byte,err error){
	getOption := getDefaultOptions()

	for _,opt := range opts{
		opt.apply(&getOption)
	}
	//fmt.Println(getOption)
	statuscode,headers,body,err = SingleReq("POST",requrl,getOption)
	return statuscode,headers,body,err
}
