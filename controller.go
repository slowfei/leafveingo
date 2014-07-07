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
//  Create on 2013-8-16
//  Update on 2014-07-02
//  Email  slowfei#foxmail.com
//  Home   http://www.slowfei.com

//
//	controller handle
//
package leafveingo

import (
	"github.com/slowfei/gosfcore/encoding/json"
	"github.com/slowfei/gosfcore/utils/filemanager"
	"github.com/slowfei/leafveingo/template"
	"net/http"
	"path"
	"reflect"
)

var (
	// controller reflect type
	RefTypeBeforeAfterController = reflect.TypeOf((*BeforeAfterController)(nil)).Elem()
	RefTypeAdeRouterController   = reflect.TypeOf((*AdeRouterController)(nil)).Elem()
)

//#pragma mark interface option ----------------------------------------------------------------------------------------------------

type ControllerOption struct {
	scheme string // "http" || "https" || ""(wildcard)
	host   string // "svn.slowfei.com" || "wwww.slowfei.com" || "slowfei.com" || ""(wildcard)
}

/**
 *	get sheme
 *
 *	@return
 */
func (c ControllerOption) Scheme() string {
	return c.scheme
}

/**
 *	set scheme
 *	"http" || "https" || ""(wildcard)
 */
func (c *ControllerOption) SetScheme(scheme string) {
	c.scheme = scheme
}

/**
 *	get host
 *
 *	@return
 */
func (c ControllerOption) Host() string {
	return c.host
}

/**
 *	set host
 *	"svn.slowfei.com" || "wwww.slowfei.com" || ""(wildcard)
 */
func (c *ControllerOption) SetHost(host string) {
	c.host = host
}

/**
 *	checked params
 */
func (c *ControllerOption) Checked() {

}

//
//	before after interface
//
type BeforeAfterController interface {

	/**
	 *  before the call controller func
	 *
	 *	@param context
	 *	@param option
	 *	@return Status200 pass, StatusNil stop Next to proceed, other status jump relevant status page
	 *
	 */
	Before(context *HttpContext, option *RouterOption) HttpStatus

	/**
	 *	after the call controller func
	 *
	 * 	@param context
	 *	@param option
	 */
	After(context *HttpContext, option *RouterOption)
}

//
//	高级路由器控制器接口，实现高级路由器机制的控制器需要实现该接口
//	特别注意，指针添加控制器和值添加的控制器的实现区别
//	指针控制器的实现：	func (c *XxxController) RouterMethodParse...
//	  值控制器的实现：func (c XxxController) RouterMethodParse...
//
type AdeRouterController interface {

	/**
	 *	路由函数解析，解析工作完全交由现实对象进行。
	 *
	 *	@param option router params option
	 *	@return funcName	is "" jump 404 page, other by name call controller
	 *	@return params		add to request form param, use set Request.URL.Query()
	 */
	RouterMethodParse(option *RouterOption) (funcName string, params map[string]string)
}

//#pragma mark Leafveingo method ----------------------------------------------------------------------------------------------------

/**
 *	controller call handle
 *
 *	@param context
 *	@param router		router interface implement object
 *	@param option		router params option
 *	@param funcName		controller call func name
 *	@param isDisp		is Dispatcher
 *	@param dispFunaName no dispatcher to nil
 */
func controllerCallHandle(context *HttpContext, router IRouter, option *RouterOption, isDisp bool, dispFunaName string) (statusCode HttpStatus, err error) {

	var funcName string = ""
	var returnValue interface{} = nil

	//	parse func name
	dispstr := ""
	if isDisp {
		dispstr = "(dispatcher)"
		funcName = dispFunaName
		statusCode = Status200
		err = nil
	} else {
		funcName, statusCode, err = router.ParseFuncName(context, option)
	}

	logInfo := "controller info" + dispstr + ": " + router.Info() + "\n"
	logInfo += "func name: [" + funcName + "]"
	context.lvServer.log.Info(logInfo)

	if Status200 != statusCode || nil != err {
		return
	}

	context.routerKeys = append(context.routerKeys, option.RouterKey)
	context.funcNames = append(context.funcNames, funcName)

	//	before
	statusCode = router.CallFuncBefore(context, option)

	if Status200 == statusCode {
		//	exce call
		returnValue, statusCode, err = router.CallFunc(context, funcName, option)

		if Status200 == statusCode && nil == err {
			if nil != returnValue {
				//	return value handle
				statusCode, err = controllerReturnValueHandle(returnValue, context, router, option, funcName)
			}
		}
	}

	//	after
	router.CallFuncAfter(context, option)

	return
}

/**
 *	controller return value handle
 *
 *	@param returnValue	body return value
 *	@param context		leafvein http context
 *	@param router		router interface
 *	@param option		router option
 *	@param funcName		request call controller func name
 *	@return statusCode	http status code
 *	@return err			error info
 */
func controllerReturnValueHandle(returnValue interface{}, context *HttpContext, router IRouter, option *RouterOption, funcName string) (statusCode HttpStatus, err error) {
	statusCode = Status200

	lv := context.lvServer
	switch cvt := returnValue.(type) {
	case string:

		context.RespWrite.Header().Set("Content-Type", "text/plain; charset="+lv.Charset())
		context.RespBodyWrite([]byte(cvt), Status200)

	case ByteOut:

		if 0 == len(cvt.ContentType) {
			context.RespWrite.Header().Set("Content-Type", "text/plain; charset="+lv.Charset())
		} else {
			context.RespWrite.Header().Set("Content-Type", cvt.ContentType)
		}
		for k, v := range cvt.Headers {
			context.RespWrite.Header().Set(k, v)
		}
		context.RespBodyWrite(cvt.Body, Status200)

	case SFJson.Json:

		context.RespWrite.Header().Set("Content-Type", "application/json; charset="+lv.Charset())
		context.RespBodyWrite(cvt.Byte(), Status200)

	case HtmlOut:

		context.RespWrite.Header().Set("Content-Type", "text/html; charset="+lv.Charset())
		context.RespBodyWrite([]byte(cvt), Status200)

	case HttpStatusValue:

		err := context.StatusPageWriteByValue(cvt)
		if nil != err {
			lv.log.Debug(err.Error())
		}

	case LVTemplate.TemplateValue:

		contentType := cvt.ContentType

		if 0 != len(contentType) {
			context.RespWrite.Header().Set("Content-Type", contentType)
		} else {
			context.RespWrite.Header().Set("Content-Type", "text/html; charset="+lv.Charset())
		}

		tplPath := cvt.TplPath
		tplName := cvt.TplName

		if 0 == len(tplPath) && 0 == len(tplName) {
			cvt.TplPath = router.ParseTemplatePath(context, funcName, option)

			if 0 == len(cvt.TplPath) {
				statusCode = Status500
				panic(ErrTemplatePathParseNil)
			}
		}

		lv.log.Info("access template path: \"" + cvt.TplPath + "\"")

		e := lv.template.Execute(context.ComperssWriter(), cvt)

		if nil != e {
			statusCode = Status500
			panic(NewLeafveinError("template: " + e.Error()))
		}

	case Redirect:

		context.RespWrite.Header().Del("Content-Encoding")
		http.Redirect(context.RespWrite, context.Request, cvt.URLPath, int(cvt.Code))
		statusCode = cvt.Code

	case Dispatcher:

		if 0 == len(cvt.FuncName) {
			statusCode = Status500
			panic(ErrControllerDispatcherFuncNameNil)
		}

		if router, ok := context.routerElement.routers[cvt.RouterKey]; ok {
			//	request的一些设置由调用者直接进行设置Header
			for k, v := range cvt.Headers {
				context.Request.Header.Set(k, v)
			}

			//	TODO 这样转发可能存在隐藏性的问题，关键在于option的作用。但目前的代码设计来说还不存在大的问题。
			option.RouterKey = cvt.RouterKey
			option.RouterPath = ""
			statusCode, err = controllerCallHandle(context, router, option, true, cvt.FuncName)

		} else {
			statusCode = Status500
			//	这个是自定义写代码的转发，如果查找不到相当于是调用者代码问题，所以直接抛出异常（恐慌）。
			dnfErr := *ErrControllerDispatcherNotFound
			dnfErr.UserInfo = "host:" + context.routerElement.host + "; router key:" + cvt.RouterKey
			panic(&dnfErr)
		}

	case ServeFilePath:
		statusCode = Status404

		filePath := string(cvt)
		if 0 != len(filePath) {

			jsonChrt := ""
			if '/' != filePath[0] {
				jsonChrt = "/"
			}

			context.RespWrite.Header().Del("Content-Encoding")

			fullPath := lv.WebRootDir() + jsonChrt + filePath

			if isExists, isDir, _ := SFFileManager.Exists(fullPath); isExists && !isDir {
				statusCode = Status200

				_, fileName := path.Split(filePath)
				context.RespWrite.Header().Set("Content-Disposition", "attachment;filename=\""+fileName+"\"")
				http.ServeFile(context.RespWrite, context.Request, fullPath)
			}
		}

	default:
		panic(ErrControllerReturnParam)
	}

	return
}
