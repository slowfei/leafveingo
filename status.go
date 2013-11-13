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
//  Create on 2013-11-13
//  Update on 2013-11-13
//  Email  slowfei@foxmail.com
//  Home   http://www.slowfei.com

//	leafveingo web 状态页处理
package leafveingo

import (
	"fmt"
	"io"
)

const (
	HttpStartTemplate = `<!doctype html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<title>Error {{status}}</title>
	<style>html{height:101%}body{padding:0;margin:0;font-size:14px;font-family:Microsoft YaHei,"微软雅黑",Lucida,Verdana,Hiragino Sans GB,STHeiti,WenQuanYi Micro Hei,SimSun,sans-serif;}.status{background-color:#000;text-align:center;padding-bottom:10px;color:#fff;}.status a{color:#FF5F00;}.status a:hover{color:#fff;}h1{padding:10px 10px;margin:0px 0px;font-size:128px;}.stack{color:red;margin:0px 30px;}.error{margin:20px 30px;font-weight:bold;color:red;}.footer{text-align:center;position:absolute;bottom:0px;height:90px;clear:both;width:100%;}.footer div{text-align:center;margin:5px 0px;}</style>
</head>
<body>
	<div class="status">
		<h1>{{status}}</h1>
		<h3>{{msg}}</h3>
		<a href="javascript:window.history.back();" title="Back to site">Back</a>
	</div>
	<div class="error">{{error}}</div>
	<div class="stack"><pre>{{stack}}</pre></div>
	<div class="footer"> 
		<div>{{Leafveingo_app_name}} Version {{Leafveingo_app_version}}</div>
		<div>Leafveingo Version {{Leafveingo_version}}</div>
		<div>Golang Version {{GoVersion}}</div>
	</div>
</body>
</html>`
)

var (
	//	http status support code
	//	http://zh.wikipedia.org/wiki/HTTP%E7%8A%B6%E6%80%81%E7%A0%81
	Status200 = HttpStatus(200)
	Status301 = HttpStatus(301)
	Status307 = HttpStatus(307)
	Status400 = HttpStatus(400)
	Status403 = HttpStatus(403)
	Status404 = HttpStatus(404)
	Status500 = HttpStatus(500)
	Status503 = HttpStatus(503)
)

//	http status code
type HttpStatus int16

//	数据的封装
type HttpStatusValue struct {
	status HttpStatus // status code
	msg    string     //
	error  string     //
	stack  string     //
}

//	new status value
func NewHttpStatusValue(status HttpStatus, msg, error, stack string) HttpStatusValue {
	return HttpStatusValue{status, msg, error, stack}
}

//	out http status page info
//	@wr
//	@value
func (lv *sfLeafvein) statusPage(wr io.Writer, value HttpStatusValue) error {

	//	根据状态代码先查找模版，直接查找模版的根目录
	tplName := fmt.Sprintf("%d%s", value.status, lv.templateSuffix)

	tmpl, err := lv.template.Parse(tplName)

	if nil != err {
		tmpl, err = lv.template.ParseString(tplName, HttpStartTemplate)

		if nil != err {
			return err
		}
	}

	data := make(map[string]string)
	data["msg"] = value.msg
	data["error"] = value.error
	data["stack"] = value.stack

	return tmpl.Execute(wr, data)
}
