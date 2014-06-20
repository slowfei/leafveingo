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
//  Update on 2013-12-31
//  Email  slowfei@foxmail.com
//  Home   http://www.slowfei.com

//	leafveingo web 的控制器操作
//
package leafveingo

import (
	"errors"
	"fmt"
	"github.com/slowfei/gosfcore/encoding/json"
	"github.com/slowfei/gosfcore/utils/filemanager"
	"github.com/slowfei/leafveingo/template"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"strings"
)

type BeforeAfterController interface {
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
 *	controller call handle
 *
 *	@param context
 *	@param routerKey	the controller router key
 *	@param funcName		controller call func name
 *	@param tplPath		template access path, no suffix
 */
func controllerCallHandle(context *HttpContext, routerKey, funcName, tplPath string) (statusCode HttpStatus, err error) {
	statusCode = Status200
	var returnValue interface{} = nil

	context.routerKeys = append(context.routerKeys, routerKey)
	context.funcNames = append(context.funcNames, funcName)

	returnValue, statusCode, err = router.CallFunc(context, funcName)

	if Status200 == statusCode && nil == err {
		if nil == returnValue {
			return
		}
		//	return value handle
		statusCode, err = controllerReturnValueHandle(returnValue, context, routerKey, funcName, tplPath)
	}

	return
}

/**
 *	controller return value handle
 *
 *	@param returnValue	call controller return value
 *	@param context
 *	@param tplPath template access path, no suffix
 */
func controllerReturnValueHandle(returnValue interface{}, context *HttpContext, routerKey, funcName, tplPath string) (statusCode HttpStatus, err error) {
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
			lvLog.Debug(err.Error())
		}

	case LVTemplate.TemplateValue:

		context.RespWrite.Header().Set("Content-Type", "text/html; charset="+lv.Charset())
		if 0 == len(cvt.TplPath) {
			if 0 != len(tplPath) {
				cvt.TplPath = tplPath + lv.templateSuffix
			} else {
				panic(ErrTemplatePathParseNil)
			}
		}

		e := lv.template.Execute(context.ComperssWriter(), cvt)

		if nil != e {
			panic(NewLeafveinError("template: " + e.Error()))
		}

	case Redirect:

		context.RespWrite.Header().Del("Content-Encoding")
		http.Redirect(context.RespWrite, context.Request, string(cvt), int(Status301))
		statusCode = Status301

	case Dispatcher:

		if 0 == len(cvt.FuncName) {
			statusCode = Status500
			panic(ErrControllerDispatcherFuncNameNil)
		}

		if router, ok := context.lvServer.routers[cvt.RouterKey]; ok {
			//	request的一些设置由调用者直接进行设置Header
			for k, v := range cvt.Headers {
				context.Request.Header.Set(k, v)
			}

			//	TODO tplPath 看看是否需要重新拼接。或则说还需要调用router.ParseController(context, option)
			//	dispatcher call
			controllerCallHandle(context, cvt.RouterKey, cvt.FuncName /*tplPath*/)

		} else {
			statusCode = Status500
			//	这个是自定义写代码的转发，如果查找不到相当于是调用者代码问题，所以直接抛出异常（恐慌）。
			err := *ErrControllerDispatcherNotFound
			err.UserInfo = "router key:" + cvt.RouterKey
			panic(&err)
		}

		if ctrlVal, ok := lv.controllers[cvt.Router]; ok {
			//	需要重新设置控制器路径，以便转发后能够查找到相应的模板
			dispCtrURLPath := strings.ToLower(cvt.Router + cvt.MethodName)

			var e error = nil
			statusCode, e = lv.cellController(cvt.Router, cvt.MethodName, "", dispCtrURLPath, context)
			if nil != e {
				lvLog.Error("dispatcher: (%v)controller (%v)method error:%v", ctrlVal.Type().String(), cvt.MethodName, e)
			}
		} else {
			statusCode = Status500
			//	这个是自定义写代码的转发，如果查找不到相当于是调用者代码问题，所以直接抛出异常（恐慌）。
			ErrControllerDispatcherNotFound.Message = lvLog.Error("dispatcher: controller not found router key:%v", cvt.Router)
			panic(ErrControllerDispatcherNotFound)
		}

	case ServeFilePath:

		context.RespWrite.Header().Del("Content-Encoding")
		filePath := path.Join(lv.WebRootDir(), string(cvt))

		if isExists, isDir, _ := SFFileManager.Exists(filePath); isExists && !isDir {
			http.ServeFile(context.RespWrite, context.Request, filePath)
		} else {
			statusCode = Status404
		}

	default:
		panic(ErrControllerReturnParam)
	}

	return
}
