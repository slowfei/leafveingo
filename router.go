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
//  Update on 2014-06-23
//  Email  slowfei#foxmail.com
//  Home   http://www.slowfei.com

//
//	router manager
//
package leafveingo

import (
	"github.com/slowfei/gosfcore/utils/strings"
)

var (
	//	globalRouterList
	_globalRouterList []globalRouter = nil
)

//
//	global router
//
type globalRouter struct {
	appName string
	router  IRouter
}

/**
 *	global add router
 *
 *	@param appName
 *	@param routerKey
 *	@param router
 */
func AddRouter(appName string, router IRouter) {
	_globalRouterList = append(_globalRouterList, globalRouter{appName, router})
}

//
//	router option
//
type RouterOption struct {
	/*
		GET http://localhost:8080/home/index.go
		routerKey 		= "/home/"
		routerPath 		= "index"
		requestMethod	= "get"
		urlSuffix       = ".go"
	*/

	routerKey     string //
	routerPath    string //	have been converted to lowercase
	requestMethod string //	GET POST ...
	urlSuffix     string //

	appName string // application name
}

//
//	router
//	TODO
//
type IRouter interface {

	/**
	 *	@return router key
	 */
	RouterKey() string

	/**
	 *	parse func name
	 *
	 *	@param context			http context
	 *	@param option			router option
	 *	@return funcName 		function name specifies call
	 *	@return statusCode		http status code, 200 pass, other to status page
	 */
	ParseFuncName(context *HttpContext, option RouterOption) (funcName string, statusCode HttpStatus, err error)

	/**
	 *	parse template path
	 *	no need to add the suffix
	 *
	 *	@param context
	 *	@param funcName	controller call func name
	 *	@return template path, suggest "[routerKey]/[funcName]"
	 */
	ParseTemplatePath(context *HttpContext, funcName string) string

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

/**
 *	router parse
 *
 *	@param context
 *	@param reqPathNoSuffix 		the no suffix request path
 *	@param reqSuffix			url request suffix
 */
func routerParse(context *HttpContext, reqPathNoSuffix, reqSuffix string) (router IRouter, option RouterOption, statusCode HttpStatus) {

	statusCode = Status404

	var lowerReqPath string
	if context.lvServer.IsReqPathIgnoreCase() {
		lowerReqPath = SFStringsUtil.ToLower(reqPathNoSuffix)
	} else {
		lowerReqPath = reqPathNoSuffix
	}

	reqPathLen := len(lowerReqPath)

	keys := context.lvServer.routerKeys
	count := len(keys)
	for i := 0; i < count; i++ {
		key := keys[i]
		keyLen := len(key)
		if keyLen <= reqPathLen && key == lowerReqPath[:keyLen] {

			if r, ok := context.lvServer.routers[key]; ok {
				option.appName = context.lvServer.AppName()
				option.urlSuffix = reqSuffix
				option.requestMethod = context.Request.Method //SFStringsUtil.ToLower(context.Request.Method)
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
