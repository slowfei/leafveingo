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
//  Update on 2014-06-20
//  Email  slowfei#foxmail.com
//  Home   http://www.slowfei.com

//	leafveingo web 状态页处理
package leafveingo

import (
	"strconv"
)

const (
	HttpStartTemplate = `<!doctype html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<title>Error {{.status}}</title>
	<style>html{height:101%}body{padding:0;margin:0;font-size:14px;font-family:Microsoft YaHei,"微软雅黑",Lucida,Verdana,Hiragino Sans GB,STHeiti,WenQuanYi Micro Hei,SimSun,sans-serif;}pre{white-space:pre-wrap;white-space:-moz-pre-wrap;white-space:-pre-wrap;white-space:-o-pre-wrap;word-wrap:break-word;font-size:14px;margin:3px 3px;}.status{background-color:#000;text-align:center;padding-bottom:10px;color:#fff;}.status a{color:#FF5F00;}.status a:hover{color:#fff;}h1{padding:10px 10px;margin:0px 0px;font-size:128px;}.stack{color:red;margin:0px 30px;}.error{margin:20px 30px;font-weight:bold;color:red;}.footer{text-align:center;position:absolute;bottom:0px;height:90px;clear:both;width:100%;}.footer div{text-align:center;margin:5px 0px;}</style>
</head>
<body>
	<div class="status">
		<h1>{{.status}}</h1>
		<h3>{{.msg}}</h3>
		<a href="javascript:window.history.back();" title="Back to site">Back</a>
	</div>
	<div class="error">{{.error}}</div>
	<div class="stack"><pre>{{.stack}}</pre></div>
	<div class="footer"> 
		<div>{{Leafveingo_app_name}} Version {{Leafveingo_app_version}}</div>
		<div>Leafveingo Version {{Leafveingo_version}}</div>
		<div>Golang Version {{GoVersion}}</div>
	</div>
</body>
</html>`

	Status301Msg = "Moved Permanently"
	Status307Msg = "Temporary Redirect"
	Status400Msg = "Bad Request"
	Status403Msg = "Forbidden The server understood the request"
	Status404Msg = "Page Not Found"
	Status500Msg = "Internal Server Error"
	Status503Msg = "Service Unavailable"
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
	status HttpStatus        // status code
	data   map[string]string //
}

//	new status value
func NewHttpStatusValue(status HttpStatus, msg, error, stack string) HttpStatusValue {

	data := make(map[string]string)
	data["msg"] = msg
	data["error"] = error
	data["stack"] = stack
	data["status"] = strconv.Itoa(int(status))

	return HttpStatusValue{status, data}
}

/**
 *	根据状态获取相应的字符串信息
 *
 *	@status
 *	@return
 */
func StatusMsg(status HttpStatus) string {
	msg := ""
	switch status {
	case Status200:
		msg = "ok"
	case Status301:
		msg = Status301Msg
	case Status307:
		msg = Status307Msg
	case Status400:
		msg = Status400Msg
	case Status403:
		msg = Status403Msg
	case Status404:
		msg = Status404Msg
	case Status500:
		msg = Status500Msg
	case Status503:
		msg = Status503Msg
	}
	return msg
}

/**
 *	status code convert string
 *
 *	@param status
 *	@return
 */
func StatusCodeToString(status HttpStatus) string {
	msg := "-1"
	switch status {
	case Status200:
		msg = "200"
	case Status301:
		msg = "301"
	case Status307:
		msg = "307"
	case Status400:
		msg = "400"
	case Status403:
		msg = "403"
	case Status404:
		msg = "404"
	case Status500:
		msg = "500"
	case Status503:
		msg = "503"
	}
	return msg
}
