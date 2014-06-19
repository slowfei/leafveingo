//	Copyright 2013 slowfei And The Contributors All rights reserved.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.
//
//  Create on 2013-9-14
//  Update on 2014-06-15
//  Email  slowfei@foxmail.com
//  Home   http://www.slowfei.com

//	路由器解析操作
//
//	目前实现的路由只是高级路由机制，还没有智能机制到。
//	解析大致流程：
//		1.分析url的后缀看是否需要去除和固定后缀访问
//		2.分析控制器的key
//			默认 "/" 则进入 默认处理函数(Index)
//			匹配router key，根据最长的key长度先匹配
//			"/admin/" -> "/auto" -> "/a" -> "/"
//			如果前面都匹配不上则最终会匹配上"/" 然后根据这个key获取相应的控制器
//			注意这个key也是作为一个目录的节点以便访问默认模板。
//			例如:"/auto" 这个就是一个目录起点，然后跟方法名进行连接
//		3.函数名解析
//			高级路由器接口实现 AdeRouterController处理：
//			只要实现了高级路由器接口，一切都以接口返回函数为标准，methodName返回长度为0则跳转404页面
//			实现接口需要返回methodName和参数map，参数map则设置到Request.Form中
//
//			默认处理规则处理：
//			1.request url = "/admin/"
//			  router key  = "/admin/" methodName = "Index"(CONTROLLER_DEFAULT_METHOD)
//			2.request url = "/about"
//			  router key  = "/"       methodName = "About"
//			3.request url = "/admin/adduser"
//			  router key  = "/admin/" methodName = "Adduser"
//
//		4.控制器与函数名的路径链接
//		  最后会将控制器key与函数名进行路径链接，以便访问默认的模板地址
//		  1.request url   "/admin-adduser"
//			router key  = "/admin" methodName = "Adduser"
//			join path   = "/admin/adduser"
//		  2.request url = "/admin/"
//			router key  = "/admin/" methodName = "Index"
//			join path   = "/admin/index"
//		  3.request url = "/"
//			router key  = "/" methodName = "Index"
//			join path   = "/index"
//
package leafveingo

import (
	"fmt"
	"github.com/slowfei/gosfcore/utils/strings"
	"path"
	"reflect"
	"regexp"
	"strings"
)

var (
	//	匹配函数名的正则，字母必须首位，后面包含0个或多个字母与数字或下划线
	_rexValidMethodName = regexp.MustCompile(`^[a-zA-Z]+[\w|\d]*$`)

	//	AdeRouterController reflect.Type
	_arcType = reflect.TypeOf((*AdeRouterController)(nil)).Elem()

	//	globalRouterList
	_globalRouterList []globalRouter = nil
)

//
//	global router
//
type globalRouter struct {
	appName   string
	routerKey string
	router    IRouter
}

/**
 *	global add router
 *
 *	@param appName
 *	@param routerKey
 *	@param router
 */
func AddRouter(appName, routerKey string, router IRouter) {
	_globalRouterList = append(_globalRouterList, globalRouter{appName, routerKey, router})
}

//
//	router option
//
type RouterOption struct {
	/*
		GET http://localhost:8080/home/Index.go
		routerKey 		= "/home/"
		routerPath 		= "index"
		requestMethod	= "get"
		urlSuffix       = ".go"
	*/

	routerKey     string // have been converted to lowercase
	routerPath    string //	lowercase
	requestMethod string // lowercase
	urlSuffix     string //

	appName string // application name
}

//
//	router
//	TODO
//
type IRouter interface {

	/**
	 *	parse controller
	 *
	 *	@param context			http context
	 *	@param option			router option
	 *	@return funcName 		function name specifies call
	 *	@return tplPath			template access path
	 *	@return statusCode		http status code, 200 pass, other to status page
	 */
	ParseController(context *HttpContext, option RouterOption) (funcName, tplPath string, statusCode HttpStatus, err error)

	/**
	 *	request func
	 *
	 *	@param context			http content
	 *	@param funcName			call controller func name
	 *	@return returnValue		controller func return value
	 *	@return statusCode		http status code, 200 pass, other to status page
	 */
	CallFunc(context *HttpContext, funcName string) (returnValue interface{}, statusCode HttpStatus, err error)

	/**
	 *	controller info
	 *
	 *	@return
	 */
	Info() string
}

//	高级路由器控制器接口，实现高级路由器机制的控制器需要实现该接口
//	特别注意，指针添加控制器和值添加的控制器的实现区别
//	指针控制器的实现：	func (c *XxxController) RouterMethodParse...
//	  值控制器的实现：func (c XxxController) RouterMethodParse...
type AdeRouterController interface {

	//	路由函数解析，解析工作完全交由现实对象进行。
	//	@requrl	请求的URL已经将后缀去除
	//	@return methodName	返回"" to 404，其他则根据返回函数名进行控制器函数的调用
	//	@return params		需要添加设置的参数，使用context.Request.URL.Query()进行设置，可返回nil。
	//
	RouterMethodParse(requrl string) (methodName string, params map[string]string)
}

/**
 *	router parse
 *
 *	@param context
 *	@param reqPathNoSuffix 		the no suffix request path
 *	@param reqSuffix			url request suffix
 */
func routerParse(context *HttpContext, reqPathNoSuffix, reqSuffix string) (router IRouter, option RouterOption, statusCode HttpStatus) {

	statusCode = Status404

	lowerReqPath := SFStringsUtil.ToLower(reqPath)
	reqPathLen := len(lowerReqPath)

	keys := context.lvServer.routerKeys
	conut := len(keys)
	for i := 0; i < count; i++ {
		key := keys[i]
		keyLen := len(key)
		if keyLen <= reqPathLen && key == lowerReqPath[:keyLen] {

			if r, ok := context.lvServer.routers[key]; ok {
				option.appName = context.lvServer.AppName()
				option.urlSuffix = reqSuffix
				option.requestMethod = SFStringsUtil.ToLower(context.Request.Method)
				option.routerKey = key
				option.routerPath = lowerReqPath[keyLen:]
				statusCode = Status200
				router = r
			} else {
				//	基本上不会进来此处
				context.lvServer.log.Error("lv.routerKeys contains %#v and lv.routers not contains %#v", key, key)
				statusCode = Status404
			}
			break
		}
	}

	return

}
