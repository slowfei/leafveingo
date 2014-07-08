package controller

import (
	"fmt"
	lv "github.com/slowfei/leafveingo"
)

//	before after 演示控制器
//	需要实现 lv.BeforeAfterController 控制器Before(...)和After(...)函数
type BeforeAfterController struct {
	tag string
}

/**
 *	before
 *
 *	@param context	固定参数
 *	@param option	固定参数
 *	@return			根据需求放回状态代码，200通过，其他将会跳转相应的状态页面，也可以返回lv.StatusNil自行响应输出。
 */
func (ba *BeforeAfterController) Before(context *lv.HttpContext, option *lv.RouterOption) lv.HttpStatus {

	fmt.Printf("BeforeAfterController(%p) Before(...)\n", ba)
	//	这里会根据请求验证密码，对了才能进入所请求的函数。
	if "123456" == context.Request.URL.Query().Get("pwd") {
		return lv.Status200
	} else {
		return lv.Status403
	}
}

/**
 *	index
 */
func (ba *BeforeAfterController) Index() string {
	fmt.Printf("BeforeAfterController(%p) Index()\n", ba)

	return "Wish you a successful visit!"
}

/**
 *	after
 *
 *	@param context	固定参数
 *	@param option	固定参数
 */
func (ba *BeforeAfterController) After(context *lv.HttpContext, option *lv.RouterOption) {
	fmt.Printf("BeforeAfterController(%p) After(...) %v\n", ba, context.Request.URL.String())
}
