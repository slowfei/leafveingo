package main

import (
	"github.com/slowfei/leafveingo"
	. "github.com/slowfei/leafveingo/example/sample/src/controllers"
)

type MainController struct {
	tag string // 主要为了区别struct不同的内存地址
}

func (m *MainController) Index() string {
	return "Hello world, Leafvingo web framework"
}

// var developer bool

func main() {

	leafvein := leafveingo.SharedLeafvein()
	// 需要在编译目录下，建立个与AppName一样的目录和webRoot目录，默认名称LeafveingoWeb
	// 具体可以看下(开发项目组织结构)
	leafvein.SetAppName("SampleWeb")
	leafvein.SetAppVersion("1.0")

	//	设置URL可访问后缀
	leafvein.SetHTTPSuffixs(".go", ".htm")

	//	设置静态文件后缀
	leafvein.SetStaticFileSuffixs(".js", ".css", ".png", ".jpg", ".gif", ".ico", ".html")

	//	设置相应输出写入是使用压缩，默认gzip优先，主要看浏览器支持的压缩类型，默认是为true的。
	leafvein.SetRespWriteCompress(true)

	//	基本主控制器访问演示
	//	http://localhost:8080/
	leafvein.AddController("/", MainController{})

	//	指针控制器演示
	//	http://localhost:8080/pointer/
	leafvein.AddController("/pointer/", &MainController{})

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

	//	开启或关闭HttpSession, 默认起始时true,这里主要是为了演示
	leafvein.SetUseSession(true)

	//	演示设置session的最大有效时间，默认30分钟
	leafvein.SetSessionMaxlifeTime(1800)

	//	演示设置session自动GC操作，自动清理session
	leafvein.SetGCSession(true)

	//	获取session manager 的演示
	// leafvein.HttpSessionManager()

	//	控制器session的演示
	//	http: //localhost:8080/s/
	leafvein.AddController("/s/", SessionController{})

	//	控制器模板演示
	//	http: //localhost:8080/t/
	leafvein.AddController("/t/", TemplateController{})

	//	设置模板目录名称，目录位于OperatingDir()下而建立的。
	//	而OperatingDi()是根据AppName建立的。
	//	模板目录名称默认就是template
	leafvein.SerTemplateDir("template")

	//	设置模板后缀，默认是.tpl
	leafvein.SetTemplateSuffix(".tpl")

	//	其他设置，以下都以默认值再设置一遍，主要为了演示。

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

	//	启动
	leafvein.Start()
}