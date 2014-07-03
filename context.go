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
//  Update on 2014-07-02
//  Email  slowfei#foxmail.com
//  Home   http://www.slowfei.com

//
//  web request context
//
package leafveingo

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"github.com/slowfei/gosfcore/utils/strings"
	"github.com/slowfei/leafveingo/session"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
)

const (
	HTML_FORM_TOKEN_NAME = "formToken"
)

//	request context
type HttpContext struct {
	lvServer        *LeafveinServer       //
	reqBody         []byte                //
	session         LVSession.HttpSession //
	reqHost         string                // request host lowercase, integrated multi-project use.
	reqScheme       string                // request scheme lowercase
	routerElement   *RouterElement        // router list element
	routerKeys      []string              // router keys
	funcNames       []string              // request controller method names
	contentEncoding string                // encoding; "none" || "gzip"...
	comperssWriter  io.Writer             //
	isCloseWriter   bool                  // is cell closeWriter()

	formparams *parampack //	form params

	RespWrite http.ResponseWriter
	Request   *http.Request
}

//	new context
func newContext(server *LeafveinServer, rw http.ResponseWriter, req *http.Request, respWirteCompress bool) *HttpContext {
	var outWrite io.Writer
	var err error

	acceptEncoding := SFStringsUtil.ToLower(req.Header.Get("Accept-Encoding"))
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

	return &HttpContext{lvServer: server, RespWrite: rw, Request: req, reqBody: nil, session: nil, contentEncoding: encoding, comperssWriter: outWrite, isCloseWriter: false}
}

/**
 *	colse comperss writer
 */
func (ctx *HttpContext) closeWriter() {
	ctx.isCloseWriter = true

	switch ow := ctx.comperssWriter.(type) {
	case *gzip.Writer:
		ow.Close()
	case *flate.Writer:
		ow.Close()
	case http.ResponseWriter:
	default:
		ctx.lvServer.log.Debug("出现未知的压缩数据类型：(%T) 可能没有进行释放.", ow)
	}
}

/**
 *	end request free context
 *
 */
func (ctx *HttpContext) free() {

	if !ctx.isCloseWriter {
		ctx.closeWriter()
	}
	ctx.lvServer = nil
	ctx.session = nil
	ctx.reqBody = nil
	ctx.comperssWriter = nil
	ctx.RespWrite = nil
	ctx.Request = nil
	ctx.routerElement = nil
}

/**
 *	get session, context the only session
 *
 *	@param resetToken is reset session token
 *	@return HttpSession
 *	@return error
 */
func (ctx *HttpContext) Session(resetToken bool) (LVSession.HttpSession, error) {

	sessionManager := ctx.lvServer.HttpSessionManager()
	if nil == sessionManager {
		return nil, ErrHttpSessionManagerClosed
	}

	if nil == ctx.session {
		var sessError error
		ctx.session, sessError = sessionManager.GetSession(ctx.RespWrite, ctx.Request, ctx.lvServer.SessionMaxlifeTime(), resetToken)
		if nil != sessError {
			ctx.lvServer.log.Error("get session error:%v", sessError)
			return nil, sessError
		}
	}
	return ctx.session, nil
}

/**
 *	get string form token
 *	Note: each access will change
 *
 *	@return
 */
func (ctx *HttpContext) FormTokenString() string {
	session, err := ctx.Session(false)
	if nil == session || nil != err {
		return ""
	}
	return session.FormTokenSignature()
}

/**
 *	get html input form token
 *	Note: each access will change
 *
 *	@return
 */
func (ctx *HttpContext) FormTokenHTML() string {
	session, err := ctx.Session(false)
	if nil == session || nil != err {
		return ""
	}
	return `<input type="hidden" name="` + HTML_FORM_TOKEN_NAME + `" value="` + session.FormTokenSignature() + `"/>`
}

/**
 *	get javascript the out form token, (主要防止页面抓取)
 *	Note: each access will change
 *
 *	@return
 */
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

/**
 *	check form token
 *
 *	@return true is pass
 */
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

/**
 *	custom check form token
 *
 *	@param true is pass
 */
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

/**
 *	get request body content
 *
 *	@return
 */
func (ctx *HttpContext) RequestBody() []byte {
	if nil == ctx.reqBody {
		readBody, err := ioutil.ReadAll(ctx.Request.Body)
		ctx.Request.Body.Close()

		if nil != err {
			ctx.lvServer.log.Error("request body get error:%v", err)
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

/**
 *	response comperss write
 *	会根据Accept-Encoding支持的格式进行压缩，优先gzip
 *
 *	@param body write content
 *	@param code	status code
 */
func (ctx *HttpContext) RespBodyWrite(body []byte, code HttpStatus) {
	ctx.RespWrite.Header().Set("Content-Encoding", ctx.contentEncoding)
	ctx.RespWrite.WriteHeader(int(code))
	ctx.comperssWriter.Write(body)
}

/**
 *	response not comperss write Content-Encoding = none
 *
 *	@param body
 *	@param code
 */
func (ctx *HttpContext) RespBodyWriteNotComperss(body []byte, code HttpStatus) {
	ctx.RespWrite.Header().Set("Content-Encoding", "none")
	ctx.RespWrite.WriteHeader(int(code))
	ctx.RespWrite.Write(body)
}

/**
 *	get comperss writer
 *
 *	@return
 */
func (ctx *HttpContext) ComperssWriter() io.Writer {
	ctx.RespWrite.Header().Set("Content-Encoding", ctx.contentEncoding)
	return ctx.comperssWriter
}

/**
 *	status page write
 *	指定模版的几个参数值，在模板中信息信息的输出
 *	模版的默认map key: {{.msg}} {{.status}} {{.error}} {{.stack}}
 *
 *	@param status status code
 *	@param msg
 *	@param error
 *	@param stack
 *	@return
 */
func (ctx *HttpContext) StatusPageWrite(status HttpStatus, msg, error, stack string) error {
	return ctx.StatusPageWriteByValue(NewHttpStatusValue(status, msg, error, stack))
}

/**
 *	status page write
 *	可以自定义指定模版的参数
 *
 *	@param value
 *	@return
 */
func (ctx *HttpContext) StatusPageWriteByValue(value HttpStatusValue) error {
	server := ctx.lvServer

	ioWr := ctx.ComperssWriter()
	ctx.RespWrite.Header().Set("Content-Type", "text/html; charset="+server.Charset())
	ctx.RespWrite.WriteHeader(int(value.status))

	err := server.statusPageExecute(value, ioWr)

	if nil != err {
		ctx.lvServer.log.Error(err.Error())
		return err
	}
	return nil
}

/**
 *	pack stauct form params
 *
 *	@param nilStruct (*TestStruct)(nil)
 *	@return new pointer struct by Type
 */
func (ctx *HttpContext) PackStructForm(nilStruct interface{}) (ptrStruct interface{}, err error) {
	refVal, e := ctx.PackStructFormByRefType(reflect.TypeOf(nilStruct))

	if nil == e {
		ptrStruct = refVal.Interface()
	} else {
		err = e
	}

	return
}

/**
 *	pack stauct form params by reflect type
 *
 *	@param structType
 *	@return refVal
 *	@return err
 */
func (ctx *HttpContext) PackStructFormByRefType(structType reflect.Type) (refVal reflect.Value, err error) {

	if nil == ctx.formparams {
		ctx.formparams, err = parampackParseForm(ctx.Request, ctx.lvServer.FileUploadSize())
		if nil != err {

			pfpErr := *ErrParampackParseFormParams
			pfpErr.UserInfo = "error info: " + err.Error()
			ctx.lvServer.log.Debug(pfpErr.Error())
			err = &pfpErr
		}
	}

	if nil == ctx.formparams {
		if nil == err {
			err = ErrParampackParseFormParams
		}
		return
	}

	//	new struct
	isNewStruct := false
	switch structType.Kind() {
	case reflect.Ptr:
		switch structType.Elem().Kind() {
		case reflect.Struct:
			refVal = parampackNewStructPtr(structType.Elem(), ctx.formparams.params)
			isNewStruct = true
		}
	case reflect.Struct:
		refVal = parampackNewStructPtr(structType, ctx.formparams.params)
		isNewStruct = true
	}

	//	set stauct field value
	if isNewStruct {
		for k, v := range ctx.formparams.params {
			count := len(v)
			switch {
			case 1 == count:
				parampackSetStructFieldValue(refVal, k, v[0])
			case 1 < count:
				parampackSetStructFieldValue(refVal, k, v)
			}

		}
		for fk, fv := range ctx.formparams.files {
			count := len(fv)
			switch {
			case 1 == count:
				parampackSetStructFieldValue(refVal, fk, fv[0])
			case 1 < count:
				parampackSetStructFieldValue(refVal, fk, fv)
			}
		}

		if reflect.Ptr != structType.Kind() {
			refVal = refVal.Elem()
		}

	} else {
		err = ErrParampackNewStruct
	}

	return
}

/**
 *	get current leafvein server
 *
 *	@return *LeafveinServer
 */
func (ctx *HttpContext) LVServer() *LeafveinServer {
	return ctx.lvServer
}

/**
 *	get router list element
 *
 *	@return *RouterElement
 */
func (ctx *HttpContext) GetIRouter(routerKey string) (IRouter, bool) {
	router, ok := ctx.routerElement.routers[routerKey]
	return router, ok
}

/**
 *	requset router keys
 *
 *	@return
 */
func (ctx *HttpContext) RouterKeys() []string {
	return ctx.routerKeys
}

/**
 *	request host
 *
 *	@return lowercase string
 */
func (ctx *HttpContext) RequestHost() string {
	return ctx.reqHost
}

/**
 *	request scheme
 *
 *	@return lowercase string
 */
func (ctx *HttpContext) RequestScheme() string {
	return ctx.reqScheme
}

/**
 *	request controller methods
 *
 *	@return
 */
func (ctx *HttpContext) FuncNames() []string {
	return ctx.funcNames
}
