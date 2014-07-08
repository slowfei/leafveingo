package controller

import (
	"fmt"
	lv "github.com/slowfei/leafveingo"
)

//
//	before after 演示控制器
//	运行时注意观察控制台打印的参数地址。
//
//	控制器分的指针传递和值传递(lv_reflect_router.go, lv_restful_router.go均已实现)
//		值传递：
//		CreateReflectController("/pointer/struct/", PointerController{})
//		每次请求(http://localhost:8080/pointer/struct/) 都会根据设置的控制器类型新建立一个对象进行处理，直到一次请求周期结束。
//
//		指针传递：
//		CreateReflectController("/pointer/", new(PointerController))
//		跟值传递相反，每次请求时都会使用设置的控制器地址进行处理，应用结束也不会改变，每次请求控制器都不会改变内存地址
//		这里涉及到并发时同时使用一个内存地址处理的问题，使用时需要注意
//
type PointerController struct {
	tag string
}

/**
 *	before
 *
 *	@param context	固定参数
 *	@param option	固定参数
 *	@return			根据需求放回状态代码，200通过，其他将会跳转相应的状态页面，也可以返回lv.StatusNil自行响应输出。
 */
func (ba *PointerController) Before(context *lv.HttpContext, option *lv.RouterOption) lv.HttpStatus {
	fmt.Printf("PointerController(%p) Before(...)\n", ba)
	return lv.Status200
}

/**
 *	index
 */
func (ba *PointerController) Index() string {
	return fmt.Sprintf("PointerController(%p) Index()", ba)
}

/**
 *	after
 *
 *	@param context	固定参数
 *	@param option	固定参数
 */
func (ba *PointerController) After(context *lv.HttpContext, option *lv.RouterOption) {
	fmt.Printf("PointerController(%p) After(...) %v\n", ba, context.Request.URL.String())
}
