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
//  Create on 2013-9-13
//  Update on 2013-10-23
//  Email  slowfei@foxmail.com
//  Home   http://www.slowfei.com

//	leafveingo web 每次请求的上下文封装
package leafveingo

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"github.com/slowfei/leafveingo/session"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	HTML_FORM_TOKEN_NAME = "formToken"
)

//	request context
type HttpContext struct {
	RespWrite       http.ResponseWriter
	Request         *http.Request
	reqBody         []byte
	session         LVSession.HttpSession
	routerKeys      []string  // router keys
	methodNames     []string  // request controller method names
	contentEncoding string    // 压缩类型存储
	comperssWriter  io.Writer //
}

//	new context
func newContext(rw http.ResponseWriter, req *http.Request, respWirteCompress bool) *HttpContext {
	var outWrite io.Writer
	var err error
	acceptEncoding := strings.ToLower(req.Header.Get("Accept-Encoding"))
	encoding := "none"

	if respWirteCompress && 0 <= strings.Index(acceptEncoding, "gzip") {
		encoding = "gzip"
		outWrite, err = gzip.NewWriterLevel(rw, gzip.BestSpeed)
	} else if respWirteCompress && 0 <= strings.Index(acceptEncoding, "deflate") {
		encoding = "deflate"
		outWrite, err = flate.NewWriter(rw, flate.BestSpeed)
	} else {
		outWrite = rw
	}

	if nil != err {
		encoding = "none"
		outWrite = rw
	}

	return &HttpContext{RespWrite: rw, Request: req, reqBody: nil, session: nil, contentEncoding: encoding, comperssWriter: outWrite}
}

// colse comperss writer
func (ctx *HttpContext) closeWriter() {
	switch ow := ctx.comperssWriter.(type) {
	case *gzip.Writer:
		ow.Close()
	case *flate.Writer:
		ow.Close()
	case http.ResponseWriter:
	default:
		lvLog.Error("出现未知的压缩数据类型：(%T) 可能没有进行释放.", ow)
	}
}

//	get session, context the only session
//	@resetToken is reset session token
func (ctx *HttpContext) Session(resetToken bool) (LVSession.HttpSession, error) {
	if nil == _thisLeafvein.HttpSessionManager() {
		return nil, ErrHttpSessionManagerClosed
	}
	if nil == ctx.session {
		var sessError error
		ctx.session, sessError = _thisLeafvein.HttpSessionManager().GetSession(ctx.RespWrite, ctx.Request, _thisLeafvein.SessionMaxlifeTime(), resetToken)
		if nil != sessError {
			lvLog.Error("get session error:%v", sessError)
			return nil, sessError
		}
	}
	return ctx.session, nil
}

//	get string form token
//	Note: each access will change
func (ctx *HttpContext) FormTokenString() string {
	session, err := ctx.Session(false)
	if nil == session || nil != err {
		return ""
	}
	return session.FormTokenSignature()
}

//	get html input form token
//	Note: each access will change
func (ctx *HttpContext) FormTokenHTML() string {
	session, err := ctx.Session(false)
	if nil == session || nil != err {
		return ""
	}
	return `<input type="hidden" name="` + HTML_FORM_TOKEN_NAME + `" value="` + session.FormTokenSignature() + `"/>`
}

//	get javascript the out form token, (主要防止页面抓取)
//	Note: each access will change
func (ctx *HttpContext) FormTokenJavascript() string {
	session, err := ctx.Session(false)
	if nil == session || nil != err {
		return ""
	}

	// `
	// <script type="text/javascript">
	// /*<![CDATA[*/
	// /***********************************************
	// * Encrypt Email script- Please keep notice intact
	// * Tool URL: http://www.dynamicdrive.com/emailriddler/
	// * **********************************************/
	// <!-- Encrypted version of: you [at] **********.*** //-->
	// var emailriddlerarray=[121,111,117,64,121,111,117,114,100,111,109,97,105,110,46,99,111,109]
	// var encryptedemail_id85='' //variable to contain encrypted email
	// for (var i=0; i<emailriddlerarray.length; i++)
	//  encryptedemail_id85+=String.fromCharCode(emailriddlerarray[i])
	// document.write('<a href="mailto:'+encryptedemail_id85+'">Contact Us</a>')
	// /*]]>*/
	// </script>
	// `

	token := session.FormTokenSignature()
	tokenByte := []byte(token)
	tblen := len(tokenByte)
	if 1 == tblen {

	}
	// 参考：http://www.dynamicdrive.com/emailriddler/
	// stript := `<script type="text/javascript"></script>`

	//	TODO 暂时还没有想到好方法实现
	return ""
}

//	check form token
func (ctx *HttpContext) CheckFormToken() bool {

	formVals := ctx.Request.Form
	if 0 == len(formVals) {
		return false
	}

	hideVal := ctx.Request.Form.Get(HTML_FORM_TOKEN_NAME)
	if 0 == len(hideVal) {
		return false
	}

	session, err := ctx.Session(false)
	if nil == session || nil != err {
		return false
	}

	return session.CheckFormTokenSignature(hideVal)

}

// custom check form token
func (ctx *HttpContext) CheckFormTokenByString(token string) bool {
	if 0 == len(token) {
		return false
	}

	session, err := ctx.Session(false)
	if nil == session || nil != err {
		return false
	}

	return session.CheckFormTokenSignature(token)
}

//	get request body content
func (ctx *HttpContext) RequestBody() []byte {
	if nil == ctx.reqBody {
		readBody, err := ioutil.ReadAll(ctx.Request.Body)
		ctx.Request.Body.Close()

		if nil != err {
			lvLog.Error("request body get error:%v", err)
			return nil
		}

		//	重新赋值body,以便下次再次读取。
		//	因为req.Body 返回的是io.ReadCloser接口，读取完后是需要Close的，但是关闭了就无法进行下次的读取了。
		//	所以就新建立个Buffer，然后使用NopCloser返回一个封装好的io.ReadCloser接口(NopCloser实现了Close函数返回nil的功能，这样代码就不需要判断太多)
		//	这样就可以无限的读取而不需要担心Close的问题了。NewBuffer是不需要操心关闭的。
		//	虽然这里nil == ctx.reqBody已经进行验证，获取一次后不会进来的，但是主要为了调用者不知道所以就NewBuffer操作。
		buf := bytes.NewBuffer(readBody)
		ctx.Request.Body = ioutil.NopCloser(buf)

		ctx.reqBody = readBody
	}
	return ctx.reqBody
}

// requset router keys
func (ctx *HttpContext) RouterKeys() []string {
	return ctx.routerKeys
}

// request controller methods
func (ctx *HttpContext) MethodNames() []string {
	return ctx.methodNames
}

//	response comperss write
// 会根据Accept-Encoding支持的格式进行压缩，优先gzip
func (ctx *HttpContext) RespBodyWrite(body []byte, code int) {
	ctx.RespWrite.Header().Set("Content-Encoding", ctx.contentEncoding)
	ctx.RespWrite.WriteHeader(code)
	ctx.comperssWriter.Write(body)
}

//  response not comperss write Content-Encoding = none
func (ctx *HttpContext) RespBodyWriteNotComperss(body []byte, code int) {
	ctx.RespWrite.Header().Set("Content-Encoding", "none")
	ctx.RespWrite.WriteHeader(code)
	ctx.RespWrite.Write(body)
}

//	get comperss writer
func (ctx *HttpContext) ComperssWriter() io.Writer {
	ctx.RespWrite.Header().Set("Content-Encoding", ctx.contentEncoding)
	return ctx.comperssWriter
}
