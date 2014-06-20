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
//  Create on 2013-8-24
//  Update on 2014-06-20
//  Email  slowfei#foxmail.com
//  Home   http://www.slowfei.com

//
//	leafveingo web response返回结果的处理
//
package leafveingo

import (
	"github.com/slowfei/gosfcore/encoding/json"
	"github.com/slowfei/leafveingo/template"
)

//	body out
type ByteOut struct {
	Body        []byte            // out content
	ContentType string            // content type "text/plain; charset=utf-8" || "image/png" ...
	Headers     map[string]string // 可以为空，看是否需要设置其他头部信息
}

//	body out HTML string
type HtmlOut string

//	redirect url
type Redirect string

//	body out file to path
type ServeFilePath string

//	dispatcher controller struct
type Dispatcher struct {
	RouterKey string
	FuncName  string
	Headers   map[string]string
}

//	out text text/plain
func BodyText(text string) string {
	return text
}

//	out html text/html
func BodyHtml(html string) HtmlOut {
	return HtmlOut(html)
}

//	out json text, application/json
func BodyJson(value interface{}) SFJson.Json {

	json, error := SFJson.NewJson(value, "", "")

	if nil != error {
		lvLog.Error("BodyJson json error:%v", error)
		json = SFJson.NewJsonNil(true)
	}

	return json
}

/**
 *	out []byte
 *
 *	@param content 	 	 body content
 * 	@param contentType 	 内容类型，"" = "text/plain; charset=utf-8"
 *	@param headers		 需要添加头部的其他信息
 *	@return ByteOut
 */
func BodyByte(content []byte, contentType string, headers map[string]string) ByteOut {
	return ByteOut{content, contentType, headers}
}

/**
 *	by template out html, text/html
 *
 *	template path: (templateDir)/(controllerRouter)/(methodName).tpl
 * 	e.g.:    URL = http://localhost:8080/index
 *	  router key = "/", funcName "Index"
 *	templatePath = (templateDir)/index.tpl
 *
 *			 URL = http://localhost:8080/Admin/index
 *	  router key = "/admin/", funcName "Index"
 *	templatePath = (templateDir)/admin/index.tpl
 *
 *	@param data
 *	@return TemplateValue
 */
func BodyTemplate(data interface{}) LVTemplate.TemplateValue {
	return LVTemplate.NewTemplateValueByData(data)
}

/**
 *	by template out html, text/html
 *	custom template path: (templateDir)/(tplPath).tpl
 *	e.g.: tplPath = "custom/base.tpl"
 *	templatePath  = (templateDir)/custom/base.tpl
 *
 *	@param tplPath custom template path
 *	@param data
 *	@return TemplateValue
 */
func BodyTemplateByTplPath(tplPath string, data interface{}) LVTemplate.TemplateValue {
	return LVTemplate.NewTemplateValue(tplPath, data)
}

/**
 *	redirect url
 *
 *	@param url
 *	@return Redirect
 */
func BodyRedirect(url string) Redirect {
	return Redirect(url)
}

/**
 *	dispatcher controller
 *
 *	@param routerKey 	router key
 *	@param funcName		controller call func name
 *	@return Dispatcher
 */
func BodyCallController(routerKey, funcName string) Dispatcher {
	return Dispatcher{routerKey, funcName, nil}
}

/**
 *	dispatcher controller, set headers info
 *
 *	@param routerKey 	router key
 *	@param funcName		controller call func name
 *	@param setHeaders	header info
 *	@return Dispatcher
 */
func BodyCallControllerByHeaders(routerKey, funcName string, setHeaders map[string]string) Dispatcher {
	return Dispatcher{routerKey, funcName, setHeaders}
}

/**
 *	out file
 *
 *	@param parh WebRootDir() path start
 *	@return ServeFilePath
 */
func BodyServeFile(path string) ServeFilePath {
	//	TODO 输出文件的保存名称默认可能是url的地址结尾名称，这个可能需要优化。
	return ServeFilePath(path)
}

/**
 *	out status page
 *
 *	@param status Status404...
 *	@param msg
 *	@param error
 *	@param stack
 *	@return HttpStatusValue
 */
func BodyStatusPage(status HttpStatus, msg, error, stack string) HttpStatusValue {
	return NewHttpStatusValue(status, msg, error, stack)
}

/**
 *	out status pata custom data map
 *
 *	@param status
 *	@param data
 *	@return HttpStatusValue
 */
func BodyStatusPageByData(status HttpStatus, data map[string]string) HttpStatusValue {
	return HttpStatusValue{status, data}
}
