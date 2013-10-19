package controller

import (
	"fmt"
	"net/http"
)

//	before after 演示控制器
type BeforeAfterController struct {
	tag string
}

//	before
//	可接收的参数：*http.Request、*url.URL、
//				*leafveingo.HttpContext、[]uint8(Request.Body)
//				http.ResponseWriter、LVSession.HttpSession
//	@return bool 返回true 可继续访问请求的函数， false 则跳转403错误页面; 默认可以设置返回值等于true
func (ba *BeforeAfterController) Before(req *http.Request, rw http.ResponseWriter) bool {
	//	这里会根据请求验证密码，对了才能进入所请求的函数。
	if "123456" == req.URL.Query().Get("pwd") {
		return true
	} else {
		return false
	}
}

// index
func (ba *BeforeAfterController) Index() string {
	return "Wish you a successful visit!"
}

//	after
//	可接收的参数：*http.Request、*url.URL、
//				*leafveingo.HttpContext、[]uint8(Request.Body)
//				http.ResponseWriter、LVSession.HttpSession
func (ba *BeforeAfterController) After(req *http.Request) {
	fmt.Printf("request end (%v) \n", req.URL.String())
}
