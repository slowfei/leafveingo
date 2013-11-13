package main

import (
	"github.com/slowfei/leafveingo"
	. "github.com/slowfei/leafveingo/example/samplebuild/src/controllers"
)

type MainController struct {
	tag string // 主要为了区别struct不同的内存地址
}

//	控制器的默认请求访问的函数(Index)，URL结尾检测为"/"( http://localhost:8080/ )
func (m *MainController) Index() string {
	return "Hello world, Leafvingo web framework"
}

//	使用配置文件加载进行配置演示
func LoadConfigInitLeafveingo(path string) leafveingo.ISFLeafvein {
	return leafveingo.InitLeafvein(path)
}

//	手动进行参数设置的演示
func BaseInitLeafveingo() leafveingo.ISFLeafvein {
	//	获取Leafveingo
	leafvein := leafveingo.SharedLeafvein()

	// 需要在项目的编译目录下，建立个与AppName一样的目录和webRoot目录，默认名称LeafveingoWeb
	// 具体可以看下(开发项目组织结构)
	leafvein.SetAppName("sample")
	leafvein.SetAppVersion("1.0")

	//	设置URL可访问后缀，可要写好了，有个"."开头
	leafvein.SetHTTPSuffixs(".go", ".htm")

	//	设置静态文件后缀
	leafvein.SetStaticFileSuffixs(".js", ".css", ".png", ".jpg", ".gif", ".ico", ".html")

	//	设置相应输出写入是使用压缩，默认gzip优先，主要看浏览器支持的压缩类型，默认是为true的。
	leafvein.SetRespWriteCompress(true)

	//	开启或关闭HttpSession, 默认起始时true,这里主要是为了演示
	leafvein.SetUseSession(true)

	//	演示设置session的最大有效时间，默认30分钟
	leafvein.SetSessionMaxlifeTime(1800)

	//	演示设置session自动GC操作，自动清理session
	leafvein.SetGCSession(true)

	//	获取session manager 的演示
	// leafvein.HttpSessionManager()

	//	设置模板目录名称，目录位于OperatingDir()下而建立的。
	//	而OperatingDir()是根据AppName()建立的。
	//	模板目录名称默认就是template
	leafvein.SerTemplateDir("template")

	//	设置模板后缀，默认是.tpl
	leafvein.SetTemplateSuffix(".tpl")

	//	其他设置，以下都以默认值再设置一遍，主要为了演示。

	//	webRoot目录
	leafvein.SetWebRootDir("webRoot")

	//	上传文件大小设置
	//	最大上传32M
	leafvein.SetFileUploadSize(32 << 20)

	//	端口设置
	leafvein.SetPort(8080)

	//	html使用的字符编码
	leafvein.SetCharset("utf-8")

	//	http 服务请求或响应超时的时间设置，秒为单位，默认0
	leafvein.SetServerTimeout(0)

	//	请求IP设置，默认为"127.0.0.1" 这样就本机访问，如果局域网或者本机IP访问就需要设置""空
	//	在服务器上，基本上是使用服务器代理进行转发，所以使用127.0.0.1，不需要IP就可以访问。
	leafvein.SetAddr("")

	return leafvein
}

func main() {

	//	使用手动设置参数信息
	// leafvein := BaseInitLeafveingo()

	//	使用配置加载
	//	"" 空 使用默认配置
	//	"sample/config/app.conf"	使用配置文件初始化
	leafvein := LoadConfigInitLeafveingo("sample/config/app.conf")

	//	基本主控制器访问演示
	//	http://localhost:8080/
	leafvein.AddController("/", MainController{})

	//	指针控制器演示
	//	http://localhost:8080/pointer/
	leafvein.AddController("/pointer/", &MainController{})

	//  原型：AddController(routerKey string, controller interface{})
	//
	//	特别说明：URL的访问规则 - AddController("router key",控制器{})
	//	http://localhost:8080/[控制器router key][控制器函数]
	//
	//	URL规则请求例子：
	//	控制器的函数名 = "User"(默认index)
	//	router  key  = "/admin"  = http://localhost:8080/adminuser
	//	router  key  = "/admin/" = http://localhost:8080/admin/user
	//
	//	控制器的函数名 = ""(默认index)ioWr
	//	router  key  = "/admin"  = http://localhost:8080/adminindex
	//	router  key  = "/admin/" = http://localhost:8080/admin/index
	//	router  key  = "/" 		 = http://localhost:8080/

	//	高级路由器演示
	//	http://localhost:8080/router/
	leafvein.AddController("/router/", RouterController{})

	//	控制器参数演示
	//	http://localhost:8080/p/
	leafvein.AddController("/p/", ParamsController{})

	//	控制器返回值演示
	//	http://localhost:8080/r/
	leafvein.AddController("/r/", ReturnParamController{})

	//	控制器Before与After函数演示
	//	http: //localhost:8080/ba/
	leafvein.AddController("/ba/", BeforeAfterController{})

	//	控制器session的演示
	//	http: //localhost:8080/s/
	leafvein.AddController("/s/", SessionController{})

	//	控制器模板演示
	//	http: //localhost:8080/t/
	leafvein.AddController("/t/", TemplateController{})

	//	状态页演示
	leafvein.AddController("/sp/", StatusController{})

	//	启动
	leafvein.Start()
}
