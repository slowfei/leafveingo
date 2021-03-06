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
//  Create on 2013-08-24
//  Update on 2015-08-03
//  Email  slowfei#nnyxing.com
//  Home   http://www.slowfei.com

//	leafveingo web 模板使用函数模块
package LVTemplate

import (
	"bytes"
	"html/template"
	"strings"
)

/**
 *	嵌入模板函数
 *
 *	@param tplPath	template relative path
 *	@param data		template data
 *	@return	template.HTML
 */
func (l *Template) embedTempate(tplPath string, data interface{}) template.HTML {

	tmpl, err := l.Parse(tplPath)

	if nil != err {
		if l.IsDevel() {
			return template.HTML(err.Error())
		} else {
			return "error code"
		}
	} else if nil != tmpl {
		buf := bytes.NewBuffer(nil)
		err = tmpl.Execute(buf, data)

		if nil != err {
			if l.IsDevel() {
				return template.HTML(err.Error())
			} else {
				return "error code"
			}
		} else {
			return template.HTML(buf.String())
		}
	}

	if l.IsDevel() {
		return "Embed Tempate Error"
	} else {
		return ""
	}
}

/**
 *	嵌入模版函数，根据模版名查询模版
 *
 *	@param tplName	cache template name
 *	@param data		template data
 *	@return	template.HTML
 */
func (l *Template) embedTempateByName(tplName string, data interface{}) template.HTML {

	if tmpl := l.goTemplate.Lookup(tplName); nil != tmpl {
		bufStr := bytes.NewBufferString("")
		tmpl.Execute(bufStr, data)
		return template.HTML(bufStr.String())
	} else {
		return template.HTML("\"" + tplName + "\" name not found template.")
	}
}

/**
 *	string转换成html标签代码
 *
 *	@param str
 *	@return
 */
func (l *Template) stringToHtml(str string) template.HTML {
	return template.HTML(str)
}

/**
 *	map类型数据的封装
 *
 *	注意，传递合并的map类型需要是map[string]interface{}，否则会出错
 *	slice值的添加："array" "value1,value2,value3"
 *				   key      value
 *
 *	@mergerMap 需要合并的map
 *	@strs	   根据key value的顺序进行添加
 */
func (l *Template) mapPack(mergerMap map[string]interface{}, strs ...string) map[string]interface{} {
	var thisMap map[string]interface{} = nil

	// TODO考虑是否需要验证mergerMap的类型

	if nil != mergerMap {
		thisMap = mergerMap
	} else {
		thisMap = make(map[string]interface{})
	}

	count := len(strs)

	for i := 0; i < count; i++ {
		if i+1 < count {
			val := strs[i+1]
			if 0 < strings.Index(val, ",") {
				slice := strings.Split(val, ",")
				thisMap[strs[i]] = slice
			} else {
				thisMap[strs[i]] = val
			}
			i++
		}
	}

	return thisMap
}
