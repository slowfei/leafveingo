package main

import (
	"github.com/slowfei/leafveingo"
	. "github.com/slowfei/leafveingo/example/samplebuild/src/controllers"
	router "github.com/slowfei/leafveingo/router"
)

type MainController struct {
	tag string // 主要为了区别struct不同的内存地址
}

//	控制器的默认请求访问的函数(Index)，URL结尾检测为"/"( http://localhost:8080/ )
func (m *MainController) Index() string {
	return "Hello world, Leafvingo web framework"
}

//	手动进行参数设置的演示
func GetLeafveinServer(option leafveingo.ServerOption) *leafveingo.LeafveinServer {

	//	new server
	//	create project directory and file:
	//		$GOPATH/samplebuild/sample				// by appName create
	//		$GOPATH/samplebuild/sample/config		// config files
	//		$GOPATH/samplebuild/sample/template		// template files
	//		$GOPATH/samplebuild/sample/webRoot		// web static file access directory
	//		$GOPATH/samplebuild/src					// source files
	//		$GOPATH/samplebuild/main.go				// main file
	//
	server := leafveingo.NewLeafveinServer("sample", option)

	//	手动设置配置
	if 0 != len(option.ConfigPath) {
		return server
	}

	//	设置app版本信息
	server.SetAppVersion("1.0")

	//	设置静态文件后缀
	server.SetStaticFileSuffixes(".js", ".css", ".png", ".jpg", ".gif", ".ico", ".html")

	//	设置相应输出写入是使用压缩，默认gzip优先，主要看浏览器支持的压缩类型，默认是为true的。
	server.SetRespWriteCompress(true)

	//	使用模版时，是否将html去除空格和换行符,默认true
	server.SetCompactHTML(true)

	//	设置请求时，是否将url忽略大小写处理。
	server.SetReqPathIgnoreCase(true)

	//	演示设置session的最大有效时间，默认30分钟
	server.SetSessionMaxlifeTime(1800)

	//	设置模板后缀，默认是.tpl
	server.SetTemplateSuffix(".tpl")

	//	上传文件大小设置
	//	最大上传32M
	server.SetFileUploadSize(32 << 20)

	//	html使用的字符编码
	server.SetCharset("utf-8")

	//	http 服务请求或响应超时的时间设置，秒为单位，默认0
	server.SetServerTimeout(0)

	return server
}

func main() {

	//	创建Server

	//	server oprion
	//
	// SetConfigPath("sample/config/app.conf") // optional default nil
	// SetAddr("127.0.0.1")                    // optional default 127.0.0.1
	// SetPort(8080)                           // optional default 8080
	// SetSMGCTime(300)                        // optional default 300, set 0 not use http session.
	option := leafveingo.DefaultOption().SetConfigPath("sample/config/app.conf").SetAddr("127.0.0.1").SetPort(8080).SetSMGCTime(300)
	server := GetLeafveinServer(option)

	//	当前主要演示反射路由，详情可以查看 https://github.com/slowfei/leafveingo/blob/master/router/lv_reflect_router.go

	//	基本主控制器访问演示
	//	http://localhost:8080/
	server.AddRouter(router.CreateReflectController("/", MainController{}))

	//	指针控制器演示
	//	http://localhost:8080/pointer/
	server.AddRouter(router.CreateReflectController("/pointer/", new(PointerController)))
	server.AddRouter(router.CreateReflectController("/pointer/struct/", PointerController{}))

	//	高级路由器演示
	//	http://localhost:8080/router/
	server.AddRouter(router.CreateReflectController("/router/", RouterController{}))

	//	控制器参数演示
	//	http://localhost:8080/p/
	server.AddRouter(router.CreateReflectController("/p/", ParamsController{}))

	//	控制器返回值演示
	//	http://localhost:8080/r/
	server.AddRouter(router.CreateReflectController("/r/", ReturnParamController{}))

	//	控制器Before与After函数演示
	//	http: //localhost:8080/ba/
	server.AddRouter(router.CreateReflectController("/ba/", BeforeAfterController{}))

	//	控制器session的演示
	//	http: //localhost:8080/s/
	server.AddRouter(router.CreateReflectController("/s/", SessionController{}))

	//	控制器模板演示
	//	http: //localhost:8080/t/
	server.AddRouter(router.CreateReflectController("/t/", TemplateController{}))

	//	状态页演示
	server.AddRouter(router.CreateReflectController("/sp/", StatusController{}))

	//	启动
	server.Start()
}
