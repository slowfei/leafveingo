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
//  Create on 2014-06-30
//  Update on 2014-07-17
//  Email  slowfei#foxmail.com
//  Home   http://www.slowfei.com

//
//	RESTful router
//
package LVRouter

import (
	"errors"
	"fmt"
	. "github.com/slowfei/leafveingo"
	"reflect"
)

//
//	RESTful controller interface
//
type RESTfulController interface {
	/**
	 *	method get
	 *
	 *	@param context
	 *	@return handle return value see response_body.go
	 */
	Get(context *HttpContext) interface{}

	/**
	 *	method post
	 *
	 *	@param context
	 *	@return
	 */
	Post(context *HttpContext) interface{}

	/**
	 *	method put
	 *
	 *	@param context
	 *	@return
	 */
	Put(context *HttpContext) interface{}

	/**
	 *	method delete
	 *
	 *	@param context
	 *	@return
	 */
	Delete(context *HttpContext) interface{}

	/**
	 *	method header
	 *
	 *	@param context
	 *	@return
	 */
	Header(context *HttpContext) interface{}

	/**
	 *	method options
	 *
	 *	@param context
	 *	@return
	 */
	Options(context *HttpContext) interface{}

	/**
	 *	other method
	 *
	 *	@param context
	 *	@return
	 */
	Other(context *HttpContext) interface{}
}

//
//	RESTful router option
//
type RESTfulRouterOption struct {
	ControllerOption
}

/**
 *	default option
 */
func DefaultRESTfulRouterOption() RESTfulRouterOption {
	option := RESTfulRouterOption{}
	option.ControllerOption.SetHost("")
	option.ControllerOption.SetScheme(URI_SCHEME_HTTP | URI_SCHEME_HTTPS)
	return option
}

/**
 *	set the support scheme
 *
 *	URI_SCHEME_HTTP | URI_SCHEME_HTTPS
 */
func (o RESTfulRouterOption) SetScheme(scheme URIScheme) RESTfulRouterOption {
	o.ControllerOption.SetScheme(scheme)
	return o
}

/**
 *	set host
 *
 *	"svn.slowfei.com" || "wwww.slowfei.com" || ""(wildcard)
 */
func (o RESTfulRouterOption) SetHost(host string) RESTfulRouterOption {
	o.ControllerOption.SetHost(host)
	return o
}

/**
 *	checked params
 */
func (o *RESTfulRouterOption) Checked() {
	o.ControllerOption.Checked()
}

//
//	RESTful router
//
//	default template parh: [host]/[routerKey]/[funcNme].[TemplateSuffix]
//						   "[host]/" multi-project use, lefveinServer.SetMultiProjectHosts("slowfei.com","svn.slowfei.com")
//	rule:
//
//		router key = "/api/object"
//			   URL = GET http://localhost:8080/api/object
//		 func name = get
//	 template path = [host]/api/object/get.tpl
//
//		router key = "/api/object"
//			   URL = POST http://localhost:8080/api/object
//		 func name = post
//	 template path = [host]/api/object/post.tpl
//
//		router key = "/api/object"
//			   URL = PUT http://localhost:8080/api/object
//		 func name = put
//	 template path = [host]/api/object/put.tpl
//
//		router key = "/api/object"
//			   URL = DELETE http://localhost:8080/api/object
//		 func name = delete
//	 template path = [host]/api/object/delete.tpl
//
//
//	url params: implement AdeRouterController interface resolve on their own
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
type RESTfulRouter struct {
	routerKey string // router key

	beforeAfter   BeforeAfterController // implement interface
	adeRouter     AdeRouterController   // implement interface
	isBeforeAfter bool                  //
	isAdeRouter   bool                  //

	controller RESTfulController //
	ctlType    reflect.Type      //

	option RESTfulRouterOption
	info   string
}

/**
 *	create RESTful router controller
 *
 *	@param routerKey	"/" || "/home/" || "/admin/"
 *	@param controller
 */
func CreateRESTfulController(routerKey string, controller interface{}) IRouter {
	return CreateRESTfulControllerWithOption(routerKey, controller, DefaultRESTfulRouterOption())
}

/**
 *	create RESTful router controller with option
 *
 *	@param option 	other params option
 */
func CreateRESTfulControllerWithOption(routerKey string, controller interface{}, option RESTfulRouterOption) IRouter {
	option.Checked()
	strBeforeAfter := ""
	strAde := ""
	ok := false

	newRefController := reflect.New(reflect.Indirect(reflect.ValueOf(controller)).Type())

	router := new(RESTfulRouter)
	router.routerKey = routerKey
	router.option = option
	router.ctlType = reflect.TypeOf(controller)
	router.controller, ok = newRefController.Interface().(RESTfulController)
	if !ok {
		panic(errors.New(fmt.Sprintf("%v does not implement RESTfulController method has pointer receiver", router.ctlType.String())))
	}

	//	使用指针类型获取所有函数，否则非指针结构获取的只能是非指针的函数
	refType := newRefController.Type()

	if refType.Implements(RefTypeAdeRouterController) {

		if reflect.Ptr == router.ctlType.Kind() {
			router.adeRouter = controller.(AdeRouterController)
		}

		router.isAdeRouter = true
		strAde = "(Implemented AdeRouterController)"
	}

	if refType.Implements(RefTypeBeforeAfterController) {

		if reflect.Ptr == router.ctlType.Kind() {
			router.beforeAfter = controller.(BeforeAfterController)
		}

		router.isBeforeAfter = true
		strBeforeAfter = "(Implemented BeforeAfterController)"
	}

	router.info = fmt.Sprintf("RESTfulRouter(%v) %v%v", router.ctlType, strBeforeAfter, strAde)

	return router
}

//# mark RESTfulRouter override IRouter -------------------------------------------------------------------------------------------

func (r *RESTfulRouter) AfterRouterParse(context *HttpContext, option *RouterOption) HttpStatus {
	statusCode := Status200

	if reflect.Ptr != r.ctlType.Kind() {
		option.RouterData = reflect.New(r.ctlType).Interface()
	}

	return statusCode
}

func (r *RESTfulRouter) ParseFuncName(context *HttpContext, option *RouterOption) (funcName string, statusCode HttpStatus, err error) {

	/* 高级路由实现操作 */
	if r.isAdeRouter {
		var params map[string]string = nil

		if nil != option.RouterData {

			adeRouter := option.RouterData.(AdeRouterController)
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

	funcName = option.RequestMethod
	statusCode = Status200

	return
}

func (r *RESTfulRouter) CallFuncBefore(context *HttpContext, option *RouterOption) HttpStatus {
	statucCode := Status200

	if r.isBeforeAfter {
		if nil != option.RouterData {

			beforeAfter := option.RouterData.(BeforeAfterController)
			statucCode = beforeAfter.Before(context, option)

		} else if nil != r.beforeAfter {
			statucCode = r.beforeAfter.Before(context, option)
		}
	}

	return statucCode
}

func (r *RESTfulRouter) CallFunc(context *HttpContext, funcName string, option *RouterOption) (returnValue interface{}, statusCode HttpStatus, err error) {

	var controller RESTfulController = nil
	if nil != option.RouterData {
		controller = option.RouterData.(RESTfulController)
	} else {
		controller = r.controller
	}

	switch funcName {
	case "get":
		returnValue = controller.Get(context)
	case "post":
		returnValue = controller.Post(context)
	case "put":
		returnValue = controller.Put(context)
	case "delete":
		returnValue = controller.Delete(context)
	case "header":
		returnValue = controller.Header(context)
	case "options":
		returnValue = controller.Options(context)
	default:
		returnValue = controller.Other(context)
	}

	statusCode = Status200

	return
}

func (r *RESTfulRouter) ParseTemplatePath(context *HttpContext, funcName string, option *RouterOption) string {

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

func (r *RESTfulRouter) CallFuncAfter(context *HttpContext, option *RouterOption) {
	if r.isBeforeAfter {
		if nil != option.RouterData {

			beforeAfter := option.RouterData.(BeforeAfterController)
			beforeAfter.After(context, option)

		} else if nil != r.beforeAfter {
			r.beforeAfter.After(context, option)
		}
	}
}

func (r *RESTfulRouter) RouterKey() string {
	return r.routerKey
}

func (r *RESTfulRouter) ControllerOption() ControllerOption {
	return r.option.ControllerOption
}

func (r *RESTfulRouter) Info() string {
	return r.info
}
