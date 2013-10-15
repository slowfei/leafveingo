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

//	路由器解析操作
//
//	目前实现的路由只是高级路由机制，还没有智能机制到。
//	解析大致流程：
//		1.分析url的后缀看是否需要去除和固定后缀访问
//		2.分析控制器的key
//			默认 "/" 则进入 默认处理函数(Index)
//			匹配router key，根据最长的key长度先匹配
//			"/admin/" -> "/auto" -> "/a" -> "/"
//			如果前面都匹配不上则最终会匹配上"/" 然后根据这个key获取相应的控制器
//			注意这个key也是作为一个目录的节点以便访问默认模板。
//			例如:"/auto" 这个就是一个目录起点，然后跟方法名进行连接
//		3.函数名解析
//			高级路由器接口实现 AdeRouterController处理：
//			只要实现了高级路由器接口，一切都以接口返回函数为标准，methodName返回长度为0则跳转404页面
//			实现接口需要返回methodName和参数map，参数map则设置到Request.Form中
//
//			默认处理规则处理：
//			1.request url = "/admin/"
//			  router key  = "/admin/" methodName = "Index"(CONTROLLER_DEFAULT_METHOD)
//			2.request url = "/about"
//			  router key  = "/"       methodName = "About"
//			3.request url = "/admin/adduser"
//			  router key  = "/admin/" methodName = "Adduser"
//
//		4.控制器与函数名的路径链接
//		  最后会将控制器key与函数名进行路径链接，以便访问默认的模板地址
//		  1.request url   "/admin-adduser"
//			router key  = "/admin" methodName = "Adduser"
//			join path   = "/admin/adduser"
//		  2.request url = "/admin/"
//			router key  = "/admin/" methodName = "Index"
//			join path   = "/admin/index"
//		  3.request url = "/"
//			router key  = "/" methodName = "Index"
//			join path   = "/index"
//
//
//	email	slowfei@foxmail.com
//	version 0.0.1.000
//	createTime 	2013-9-14
//	updateTime	2013-10-10
package leafveingo

import (
	"github.com/slowfei/gosfcore/log"
	"path"
	"reflect"
	"regexp"
	"strings"
)

var (
	//	匹配函数名的正则，字母必须首位，后面包含0个或多个字母与数字或下划线
	_rexValidMethodName = regexp.MustCompile(`^[a-zA-Z]+[\w|\d]*$`)

	//	AdeRouterController reflect.Type
	_arcType = reflect.TypeOf((*AdeRouterController)(nil)).Elem()
)

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

//	url路由器解析
//
func (lv *sfLeafvein) routerParse(reqPath string, context *HttpContext) (routerKey, methodName, ctrlPath string, stuctCode int) {
	SFLog.Info(" request url path: %#v", reqPath)

	//	去除url请求后缀
	//	/index.go = /index
	if 0 < len(lv.suffixs) && '/' != reqPath[len(reqPath)-1] {
		//	固定后缀的访问
		isSuffix := false
		for _, suffix := range lv.suffixs {
			if strings.HasSuffix(reqPath, suffix) {
				reqPath = reqPath[:len(reqPath)-len(suffix)]
				isSuffix = true
				break
			}
			if 0 == len(suffix) {
				isSuffix = true
				break
			}
		}
		if !isSuffix {
			//	跳转404页面
			stuctCode = HTTP_STATUS_CODE_404
			SFLog.Info("Invalid suffix")
			return
		}

	} else {
		//	检测是否有后缀名，有则去除
		suffixIndex := strings.LastIndex(reqPath, ".")
		if 0 < suffixIndex {
			reqPath = reqPath[:suffixIndex]
		}

	}

	//	验证是否调用控制器解析路由函数方法成功，不成功则调用默认处理机制。
	isCallSuccess := true

	//	路由器key解析，根据请求的url解析出添加控制器时的router key和函数名
	if "/" == reqPath {
		//	选择默认的控制器
		routerKey = "/"
	} else {
		//	遍历添加控制器时设置的router key，匹配是否符合，只要与url前段匹配得上就说明调用该控制器
		//	controllerKeys已经根据长度由长到短进行排序，完全匹配不了会进入"/"主控制器
		compareReqPath := strings.ToLower(reqPath)
		for _, rk := range lv.controllerKeys {
			if 0 == strings.Index(compareReqPath, rk) {
				//	methodName为截取控制器key的后段部分
				methodName = reqPath[len(rk):]
				routerKey = rk
				break
			}
		}

		//	如果methodName正好被截取为0，表示需要进入的是控制器默认方法，所以可以直接执行默认的处理机制
		if 0 != len(methodName) {
			if arcImpl, ok := lv.controllerArcImpls[routerKey]; ok {
				var params map[string]string = nil
				methodName, params = arcImpl.RouterMethodParse(reqPath)

				if 0 == len(methodName) {
					//	404
					stuctCode = HTTP_STATUS_CODE_404
					return
				}

				if 0 != len(params) {
					values := context.Request.URL.Query()
					for k, v := range params {
						values.Set(k, v)
					}
					context.Request.URL.RawQuery = values.Encode()
				}
				isCallSuccess = false
			}
		}
	}

	//	执行函数名的默认处理机制
	if isCallSuccess {
		if 0 == len(methodName) {
			methodName = CONTROLLER_DEFAULT_METHOD
		} else {
			//	匹配函数名是否正确
			if !_rexValidMethodName.MatchString(methodName) {
				//	404
				stuctCode = HTTP_STATUS_CODE_404
				return
			} else {
				//	控制器的函数名称将首字母转换成大写
				methodName = strings.Title(methodName)

			}
		}

	}

	//	控制器与函数名链接的路径，主要为了能够默认进入模板路径文件
	ctrlPath = path.Join(routerKey, strings.ToLower(methodName))

	if "GET" != context.Request.Method {
		//	每个不同的method进行不同的函数调用
		//	例如post = PostAbout
		method := context.Request.Method
		methodName = strings.Title(strings.ToLower(method)) + methodName
	}

	SFLog.Info("controller   key: %#v   method name: %#v   path: %#v \n", reqPath, methodName, ctrlPath)

	stuctCode = HTTP_STATUS_CODE_200
	return

}
