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
//  Update on 2013-10-23
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

//	控制器返回参数处理
func (lv *sfLeafvein) returnValue(v []reflect.Value, ctrURLPath string, context *HttpContext) (stuctCode HttpStatus, err error) {
	stuctCode = Status200

	if 1 == len(v) {
		rv := reflect.Indirect(v[0])
		// rw.Header().Set("Content-Length", "10000")
		switch cvt := rv.Interface().(type) {
		case string:
			context.RespWrite.Header().Set("Content-Type", "text/plain; charset="+lv.charset)
			context.RespBodyWrite([]byte(cvt), Status200)
		case ByteOut:
			if 0 == len(cvt.ContentType) {
				context.RespWrite.Header().Set("Content-Type", "text/plain; charset="+lv.charset)
			} else {
				context.RespWrite.Header().Set("Content-Type", cvt.ContentType)
			}
			for k, v := range cvt.Headers {
				context.RespWrite.Header().Set(k, v)
			}
			context.RespBodyWrite(cvt.Body, Status200)
		case SFJson.Json:
			context.RespWrite.Header().Set("Content-Type", "application/json")
			context.RespBodyWrite(cvt.Byte(), Status200)
		case HtmlOut:
			context.RespWrite.Header().Set("Content-Type", "text/html; charset="+lv.charset)
			context.RespBodyWrite([]byte(cvt), Status200)
		case LVTemplate.TemplateValue:
			context.RespWrite.Header().Set("Content-Type", "text/html; charset="+lv.charset)
			if "" == cvt.TplPath {
				cvt.TplPath = ctrURLPath + lv.templateSuffix
			}
			e := lv.template.Execute(context.ComperssWriter(), cvt)
			if nil != e {
				panic(e.Error())
			}
		case Redirect:
			context.RespWrite.Header().Del("Content-Encoding")
			http.Redirect(context.RespWrite, context.Request, string(cvt), int(Status301))
			stuctCode = Status301
		case Dispatcher:
			if 0 == len(cvt.MethodName) {
				cvt.MethodName = CONTROLLER_DEFAULT_METHOD
			}

			if ctrlVal, ok := lv.controllers[cvt.Router]; ok {
				//	需要重新设置控制器路径，以便转发后能够查找到相应的模板
				dispCtrURLPath := strings.ToLower(cvt.Router + cvt.MethodName)

				//	request的一些设置由调用者直接进行设置Header
				for k, v := range cvt.Headers {
					context.Request.Header.Set(k, v)
				}

				var e error = nil
				stuctCode, e = lv.cellController(cvt.Router, cvt.MethodName, dispCtrURLPath, context)
				if nil != e {
					lvLog.Error("dispatcher: (%v)controller (%v)method error:%v", ctrlVal.Type().String(), cvt.MethodName, e)
				}
			} else {
				stuctCode = Status500
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
				//	404
				http.NotFound(context.RespWrite, context.Request)
			}

		default:
			panic(ErrControllerReturnParam)
		}

	} else if 1 < len(v) {
		panic(ErrControllerReturnParamNum)
	}

	return
}

//	解析http提交的参数，包括上传文件的信息
func (lv *sfLeafvein) parseFormParams(req *http.Request) (
	urlValues url.Values,
	files map[string][]*multipart.FileHeader,
	fileNum int,
	err error,
	stuctCode HttpStatus) {

	switch req.Method {
	case "GET":
		urlValues = req.URL.Query()
	case "POST":
		contentType := req.Header.Get("Content-Type")
		enctype, _, e := mime.ParseMediaType(contentType)
		if nil != e {
			err = e
			stuctCode = Status400
			return
		}

		switch {
		case enctype == "application/x-www-form-urlencoded":
			err = req.ParseForm()
			if nil != err {
				stuctCode = Status400
				return
			}

			//	考虑安全因素，让调用则知道请求的参数来自Form还是Query所以进行POST请求就只获取Form的参数
			urlValues = req.PostForm

		case enctype == "multipart/form-data":
			err = req.ParseMultipartForm(lv.fileUploadSize)
			if nil != err {
				stuctCode = Status400
				return
			}
			// ParseMultipartForm()解析已经调用了ParseForm()
			urlValues = url.Values(req.MultipartForm.Value)

			if 0 < len(req.MultipartForm.File) {
				for k, v := range req.MultipartForm.File {
					//	添加空字符串的主要目的是为了能够在创建结构时初始化切片的数量
					urlValues.Set(k, "")
					fNum := len(v)
					if 1 < fNum {
						fileNum += fNum
					} else {
						fileNum++
					}
				}
				files = req.MultipartForm.File
			}

		}
	default:
		urlValues = make(url.Values)
	}

	stuctCode = Status200
	return
}

//	操作控制器请求的函数
//	@isPackStruct 是否封装用户自定义的struct
func (lv *sfLeafvein) cellMethod(
	methodValue reflect.Value,
	urlValues url.Values,
	files map[string][]*multipart.FileHeader,
	context *HttpContext,
	isPackStruct bool) []reflect.Value {

	//	内部使用的结构设值
	setStructValue := func(refValue reflect.Value) {
		for k, v := range urlValues {
			count := len(v)
			switch {
			case 1 == count:
				lv.setStructFieldValue(refValue, k, v[0])
			case 1 < count:
				lv.setStructFieldValue(refValue, k, v)
			}

		}
		for fk, fv := range files {
			count := len(fv)
			switch {
			case 1 == count:
				lv.setStructFieldValue(refValue, fk, fv[0])
			case 1 < count:
				lv.setStructFieldValue(refValue, fk, fv)
			}
		}
	}

	methodType := methodValue.Type()
	argsNum := methodType.NumIn()
	args := make([]reflect.Value, argsNum, argsNum)

	for i := 0; i < argsNum; i++ {
		in := methodType.In(i)
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
			switch in.Kind() {
			case reflect.Ptr:
				switch in.Elem().Kind() {
				case reflect.Struct:
					if isPackStruct {
						newValue := lv.newStructPtr(in.Elem(), urlValues)
						setStructValue(newValue)
						argsValue = newValue
					}
				}
			case reflect.Struct:
				if isPackStruct {
					newValue := lv.newStructPtr(in, urlValues)
					setStructValue(newValue)
					argsValue = newValue.Elem()
				}
			}
		}

		if reflect.Invalid == argsValue.Kind() {
			argsValue = reflect.Zero(in)
		}

		args[i] = argsValue
	}
	return methodValue.Call(args)
}

//	操作控制器请求的处理
//	@controller		请求的控制器
//	@funcName		控制器函数名
//	@ctrURLPath		控制器url路径
//	@rw
//	@req
func (lv *sfLeafvein) cellController(routerKey, methodName, ctrlPath string, context *HttpContext) (stuctCode HttpStatus, err error) {
	ctrlVal, ok := lv.controllers[routerKey]

	if !ok {
		err = errors.New(lvLog.Error("cellController not found router key:%v", routerKey))
		stuctCode = Status404
		return
	}

	//	TODO 考虑看是否真的需要使用slice来进行存储。
	context.routerKeys = append(context.routerKeys, routerKey)
	context.methodNames = append(context.methodNames, methodName)

	switch ctrlVal.Kind() {
	case reflect.Ptr:
		//	指针类型不做操作，直接使用该对象。
	default:
		//	如果判断不为指针结构的值，就创建一个新的
		ctrlVal = reflect.New(ctrlVal.Type())
	}

	//	控制器调用方法
	var requestMethodValue reflect.Value

	//	控制器调用前和后处理的函数
	var beforeMethodValue reflect.Value
	isBefore := false
	var afterMethodValue reflect.Value
	isAfter := false

	//	是否执行调用控制器方法
	isCell := false

	if requestMethodValue = ctrlVal.MethodByName(methodName); requestMethodValue.IsValid() {
		isCell = true

		if beforeMethodValue = ctrlVal.MethodByName(CONTROLLER_BEFORE_METHOD); beforeMethodValue.IsValid() {
			isBefore = true
		}
		if afterMethodValue = ctrlVal.MethodByName(CONTROLLER_AFTER_METHOD); afterMethodValue.IsValid() {
			isAfter = true
		}

	} else {
		err = errors.New(fmt.Sprintf("cellController not found method name:%v", methodName))
		stuctCode = Status404
		return
	}

	if isCell {

		//	请求参数
		var urlValues url.Values
		//	上传文件
		var files map[string][]*multipart.FileHeader
		//	记录文件请求数量
		fileNum := 0

		urlValues, files, fileNum, err, stuctCode = lv.parseFormParams(context.Request)
		if Status200 != stuctCode {
			return
		}

		logInfo := fmt.Sprintf("Request controller: %s \nRequest param: %v\nRequest fileNum: %d\n", ctrlVal.Type(), urlValues, fileNum)
		lvLog.Info(logInfo)

		isCellCtrMethod := true
		if isBefore {
			returnValues := lv.cellMethod(beforeMethodValue, urlValues, files, context, false)
			if 1 == len(returnValues) {
				retVal := returnValues[0]
				if retVal.Kind() == reflect.Bool {
					isCellCtrMethod = retVal.Bool()
				}
			}
		}

		if isCellCtrMethod {
			//	request controller cell method
			rvs := lv.cellMethod(requestMethodValue, urlValues, files, context, true)
			//	处理控制器返回值
			stuctCode, err = lv.returnValue(rvs, ctrlPath, context)
		} else {
			//	请求拒绝
			stuctCode = Status403
			//	error info is return?
		}

		if isAfter {
			lv.cellMethod(afterMethodValue, urlValues, files, context, false)
		}

	}
	return
}
