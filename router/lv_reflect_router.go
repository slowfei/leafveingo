//	Copyright 2014 slowfei And The Contributors All rights reserved.
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
//  Create on 2014-06-16
//  Update on 2014-07-17
//  Email  slowfei#foxmail.com
//  Home   http://www.slowfei.com

//
//	reflect router
//
package LVRouter

import (
	"fmt"
	. "github.com/slowfei/leafveingo"
	"reflect"
)

var (
	//	controller default url request ("http://localhost:8080/") method
	CONTROLLER_DEFAULT_METHOD = "Index"
)

//
//	reflect router other option
//
type ReflectRouterOption struct {
	ControllerOption
}

/**
 *	default option
 */
func DefaultReflectRouterOption() ReflectRouterOption {
	option := ReflectRouterOption{}
	option.ControllerOption.SetHost("")
	option.ControllerOption.SetScheme(URI_SCHEME_HTTP | URI_SCHEME_HTTPS)
	return option
}

/**
 *	set the support scheme
 *
 *	URI_SCHEME_HTTP | URI_SCHEME_HTTPS
 */
func (o ReflectRouterOption) SetScheme(scheme URIScheme) ReflectRouterOption {
	o.ControllerOption.SetScheme(scheme)
	return o
}

/**
 *	set host
 *	"svn.slowfei.com" || "wwww.slowfei.com" || ""(wildcard)
 */
func (o ReflectRouterOption) SetHost(host string) ReflectRouterOption {
	o.ControllerOption.SetHost(host)
	return o
}

/**
 *	checked params
 */
func (o *ReflectRouterOption) Checked() {
	o.ControllerOption.Checked()
}

//
//	reflect router
//
//	default template parh: [host]/[routerKey]/[funcNme].[TemplateSuffix]
//						   "[host]/" multi-project use, lefveinServer.SetMultiProjectHosts("slowfei.com","svn.slowfei.com")
//	rule:
//
//		router key = "/"
//			   URL = GET http://localhost:8080/
//		 func name = Index
//	 template path = [host]/Index.tpl
//
//		router key = "/"
//			   URL = POST http://localhost:8080/
//		 func name = PostIndex
//	 template path = [host]/PostIndex.tpl
//
//		router key = "/"
//			   URL = Get http://localhost:8080/user#!list
//		 func name = UserList
//	 template path = [host]/UserList.tpl
//
//		router key = "/"
//			   URL = Post http://localhost:8080/user[^a-zA-Z]+list[^a-zA-Z]+auto
//		 func name = PostUserListAuto
//	 template path = [host]/PostUserListAuto.tpl
//
//		router key = "/admin/"
//			   URL = GET http://localhost:8080/admin/login
//		 func name = Login
//	 template path = [host]/admin/Login.tpl
//
//		router key = "/admin/"
//			   URL = POST http://localhost:8080/admin/login
//		 func name = PostLogin
//	 template path = [host]/admin/PostLogin.tpl
//
//
//	控制器分的指针传递和值传递
//		值传递：
//		CreateReflectController("/pointer/struct/", PointerController{})
//		每次请求(http://localhost:8080/pointer/struct/) 都会根据设置的控制器类型新建立一个对象进行处理，直到一次请求周期结束。
//
//		指针传递：
//		CreateReflectController("/pointer/", new(PointerController))
//		跟值传递相反，每次请求时都会使用设置的控制器地址进行处理，应用结束也不会改变，每次请求控制器都不会改变内存地址
//		这里涉及到并发时同时使用一个内存地址处理的问题，使用时需要注意
//
type ReflectRouter struct {
	routerKey string // router key

	beforeAfter   BeforeAfterController // implement interface
	adeRouter     AdeRouterController   // implement interface
	isBeforeAfter bool
	isAdeRouter   bool

	ctlRefVal     reflect.Value  // controller reflect value
	checkFuncName map[string]int // check func name map

	option  ReflectRouterOption
	info    string
	typestr string
}

/**
 *	create reflect router controller
 *
 *	@param routerKey	"/" || "/home/" || "/admin/"
 *	@param controller
 */
func CreateReflectController(routerKey string, controller interface{}) IRouter {
	return CreateReflectControllerWithOption(routerKey, controller, DefaultReflectRouterOption())
}

/**
 *	create reflect router controller with option
 *
 *	@param option 	other params option
 */
func CreateReflectControllerWithOption(routerKey string, controller interface{}, option ReflectRouterOption) IRouter {
	option.Checked()
	strBeforeAfter := ""
	strAde := ""

	refRouter := new(ReflectRouter)
	refRouter.routerKey = routerKey
	refRouter.checkFuncName = make(map[string]int)
	refRouter.ctlRefVal = reflect.ValueOf(controller)
	refRouter.option = option
	refRouter.isAdeRouter = false
	refRouter.isBeforeAfter = false

	//	使用指针类型获取所有函数，否则非指针结构获取的只能是非指针的函数
	refType := reflect.New(reflect.Indirect(refRouter.ctlRefVal).Type()).Type()

	if refType.Implements(RefTypeAdeRouterController) {

		if reflect.Ptr == refRouter.ctlRefVal.Kind() {
			refRouter.adeRouter = controller.(AdeRouterController)
		}

		refRouter.isAdeRouter = true
		strAde = "(Implemented AdeRouterController)"
	}

	if refType.Implements(RefTypeBeforeAfterController) {

		if reflect.Ptr == refRouter.ctlRefVal.Kind() {
			refRouter.beforeAfter = controller.(BeforeAfterController)
		}

		refRouter.isBeforeAfter = true
		strBeforeAfter = "(Implemented BeforeAfterController)"
	}

	for i := 0; i < refType.NumMethod(); i++ {
		refMet := refType.Method(i)
		funcName := refMet.Name
		if funcName[0] >= 'A' && funcName[0] <= 'Z' {
			refRouter.checkFuncName[funcName] = i
		}
	}

	refRouter.typestr = refRouter.ctlRefVal.Type().String()
	refRouter.info = fmt.Sprintf("ReflectRouter(%v) %v%v", refRouter.ctlRefVal.Type(), strBeforeAfter, strAde)

	return refRouter
}

/**
 *	func name suffix handle
 *
 *	@param funcName
 *	@param option
 */
func (r *ReflectRouter) funcNameSuffixHandle(funcName string, option *RouterOption) string {
	result := funcName

	urlSuffix := option.UrlSuffix

	if 0 != len(urlSuffix) {
		firstc := urlSuffix[0]
		if firstc >= 'a' && firstc <= 'z' {
			firstc -= 'a' - 'A'
		}
		first := string(firstc)
		urlSuffix = first + urlSuffix[1:]

		tempFunc := funcName + urlSuffix

		if _, ok := r.checkFuncName[tempFunc]; ok {
			result = tempFunc
		}
	}

	return result
}

/**
 *	get func params
 *
 *	@param funcType
 *	@param
 */
func (r *ReflectRouter) getFuncArgs(funcType reflect.Type, context *HttpContext) []reflect.Value {
	argsNum := funcType.NumIn()
	args := make([]reflect.Value, argsNum, argsNum)

	for i := 0; i < argsNum; i++ {
		in := funcType.In(i)
		typeString := in.String()
		var argsValue reflect.Value

		switch typeString {
		case "*http.Request":
			argsValue = reflect.ValueOf(context.Request)
		case "http.Request":
			argsValue = reflect.ValueOf(context.Request).Elem()
		case "*url.URL":
			argsValue = reflect.ValueOf(context.Request.URL)
		case "url.URL":
			argsValue = reflect.ValueOf(context.Request.URL).Elem()
		case "*leafveingo.HttpContext":
			argsValue = reflect.ValueOf(context)
		case "leafveingo.HttpContext":
			argsValue = reflect.ValueOf(context).Elem()
		case "[]uint8":
			body := context.RequestBody()
			if nil != body {
				argsValue = reflect.ValueOf(body)
			} else {
				argsValue = reflect.Zero(in)
			}
		case "http.ResponseWriter":
			argsValue = reflect.ValueOf(context.RespWrite)
		case "LVSession.HttpSession":
			session, _ := context.Session(false)
			if nil != session {
				argsValue = reflect.ValueOf(session)
			} else {
				argsValue = reflect.Zero(in)
			}
		default:
			val, err := context.PackStructFormByRefType(in)
			if nil == err {
				argsValue = val
			} else {
				context.LVServer().Log().Debug(err.Error())
			}
		}

		if reflect.Invalid == argsValue.Kind() {
			argsValue = reflect.Zero(in)
		}

		args[i] = argsValue
	}

	return args
}

//# mark ReflectRouter override IRouter -------------------------------------------------------------------------------------------

func (r *ReflectRouter) AfterRouterParse(context *HttpContext, option *RouterOption) HttpStatus {
	statusCode := Status200

	if reflect.Ptr != r.ctlRefVal.Kind() {
		option.RouterDataRefVal = reflect.New(r.ctlRefVal.Type())
	}

	return statusCode
}

func (r *ReflectRouter) ParseFuncName(context *HttpContext, option *RouterOption) (funcName string, statusCode HttpStatus, err error) {

	/* 高级路由实现操作 */
	if r.isAdeRouter {
		var params map[string]string = nil

		if reflect.Invalid != option.RouterDataRefVal.Kind() {

			adeRouter := option.RouterDataRefVal.Interface().(AdeRouterController)
			funcName, params = adeRouter.RouterMethodParse(option)

		} else if nil != r.adeRouter {
			funcName, params = r.adeRouter.RouterMethodParse(option)
		}

		if 0 == len(funcName) {
			statusCode = Status404
		} else {
			statusCode = Status200
		}

		if 0 != len(params) {
			values := context.Request.URL.Query()
			for k, v := range params {
				values.Set(k, v)
			}
			context.Request.URL.RawQuery = values.Encode()
		}
		return
	}

	statusCode = Status404
	method := option.RequestMethod
	reqPath := option.RouterPath

	/* parse func name prefix */
	funcNamePrefix := ""
	if "get" != method {
		firstc := method[0]
		if firstc >= 'a' && firstc <= 'z' {
			firstc -= 'a' - 'A'
		}
		first := string(firstc)
		funcNamePrefix = first + method[1:]
	}

	/* parse func name */

	//	url = "http://localhost:8080/router/" router key = "/router/" || "/router"
	//	reqPath = "" || "/" to Default func name
	if 0 == len(reqPath) || (1 == len(reqPath) && '/' == reqPath[0]) {
		statusCode = Status200
		funcName = r.funcNameSuffixHandle(funcNamePrefix+CONTROLLER_DEFAULT_METHOD, option)
		return
	}

	//	url = "http://localhost:8080/router/[reqPath]" router key = "/router/"
	//	reqPath = "list" funcName = "List"
	//	reqPath = "list#!json" || "list[^a-zA-Z]*json" funcName = "ListJson"
	//	reqPath = "list/user"  funcName = "ListUser"
	//	reqPath = "list/user/auto" || "list[^a-zA-Z]+user[^a-zA-Z]+auto"  funcName = "ListUserAuto"

	nameByte := make([]byte, len(reqPath))
	isUpper := true
	writeIdx := 0

	count := len(reqPath)
	for i := 0; i < count; i++ {
		c := reqPath[i]

		AZ := c >= 'A' && c <= 'Z'
		az := c >= 'a' && c <= 'z'
		number := c >= '0' && c <= '9'

		if AZ || az || number {
			if isUpper {
				isUpper = false
				if az {
					c -= 'a' - 'A'
				}
			}
			nameByte[writeIdx] = c
			writeIdx++
		} else {
			isUpper = true
		}

	}

	if 0 != writeIdx {
		funcName = r.funcNameSuffixHandle(funcNamePrefix+string(nameByte[:writeIdx]), option)
		statusCode = Status200
	} else {
		statusCode = Status404
	}

	return
}

func (r *ReflectRouter) CallFuncBefore(context *HttpContext, option *RouterOption) HttpStatus {
	statucCode := Status200

	if r.isBeforeAfter {

		if reflect.Invalid != option.RouterDataRefVal.Kind() {

			beforeAfter := option.RouterDataRefVal.Interface().(BeforeAfterController)
			statucCode = beforeAfter.Before(context, option)

		} else if nil != r.beforeAfter {
			statucCode = r.beforeAfter.Before(context, option)
		}
	}

	return statucCode
}

func (r *ReflectRouter) CallFunc(context *HttpContext, funcName string, option *RouterOption) (returnValue interface{}, statusCode HttpStatus, err error) {

	if index, ok := r.checkFuncName[funcName]; ok {
		statusCode = Status200

		var controller reflect.Value

		if reflect.Invalid != option.RouterDataRefVal.Kind() {
			controller = option.RouterDataRefVal
		} else {
			controller = r.ctlRefVal
		}

		refMet := controller.Method(index)

		//	get params
		args := r.getFuncArgs(refMet.Type(), context)

		//	call method
		reVals := refMet.Call(args)

		if 0 != len(reVals) {
			returnValue = reVals[0].Interface()
		}

	} else {
		statusCode = Status404
		err = NewLeafveinError("(" + r.typestr + ") not found func name: " + funcName)
	}

	return
}

func (r *ReflectRouter) ParseTemplatePath(context *HttpContext, funcName string, option *RouterOption) string {

	if 0 == len(funcName) {
		return ""
	}

	path := r.routerKey
	pathLen := len(path)
	name := funcName
	nameLen := len(name)

	if '/' == path[0] {
		path = path[1:]
		pathLen = len(path)
	}
	if '/' == path[pathLen-1] {
		path = path[:pathLen-1]
	}
	if '/' == name[0] {
		name = name[1:]
		nameLen = len(name)
	}
	if '/' == name[nameLen-1] {
		name = name[:nameLen-1]
	}

	hostPath := ""
	host := context.RequestHost()
	if 0 != len(host) {
		hostPath = host + "/"
	}

	return hostPath + path + "/" + name + context.LVServer().TemplateSuffix()
}

func (r *ReflectRouter) CallFuncAfter(context *HttpContext, option *RouterOption) {
	if r.isBeforeAfter {
		if reflect.Invalid != option.RouterDataRefVal.Kind() {

			beforeAfter := option.RouterDataRefVal.Interface().(BeforeAfterController)
			beforeAfter.After(context, option)

		} else if nil != r.beforeAfter {
			r.beforeAfter.After(context, option)
		}
	}
}

func (r *ReflectRouter) RouterKey() string {
	return r.routerKey
}

func (r *ReflectRouter) ControllerOption() ControllerOption {
	return r.option.ControllerOption
}

func (r *ReflectRouter) Info() string {
	return r.info
}
