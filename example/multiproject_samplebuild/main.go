package main

import (
	"github.com/slowfei/leafveingo"
	"github.com/slowfei/leafveingo/example/multiproject_samplebuild/src/controllers/admin"
	"github.com/slowfei/leafveingo/example/multiproject_samplebuild/src/controllers/blog"
	"github.com/slowfei/leafveingo/example/multiproject_samplebuild/src/controllers/www"
	router "github.com/slowfei/leafveingo/router"
)

func main() {
	//	配置日志管理，这个需要手动配置
	leafveingo.SetLogManager(leafveingo.DEFAULT_LOG_CHANNEL_SIZE)

	/*
		---------------------------------------------------------------------------------
		多项目继承的演示
		1. 同一个端口配置不同的host(www.slowfei.com||blog.slowfei.com)进行访问
		2. admin后台管理 配置另一个端口进行访问
		3. https 演示操作

			   编译 go build
		开发模式启动 ./multiproject_samplebuild -devel
		启动后浏览器打开地址：http://localhost:8080/index.html
		---------------------------------------------------------------------------------
	*/

	/*
		-----------------
		slowfei.com and blog.slowfei.com server
	*/
	option := leafveingo.DefaultOption().SetConfigPath("multiproject/config/app.conf").SetAddr("127.0.0.1").SetPort(8080).SetSMGCTime(300)
	server := leafveingo.NewLeafveinServer("multiproject", option)

	//	添加多项目的支持，在同一个端口下
	//	create project directory and file:
	//		$GOPATH/multiproject_samplebuild/multiproject								// by appName create
	//		$GOPATH/multiproject_samplebuild/multiproject/config						// config files
	//		$GOPATH/multiproject_samplebuild/multiproject/template/slowfei.com			// template files
	//		$GOPATH/multiproject_samplebuild/multiproject/template/blog.slowfei.com		// template files
	//		$GOPATH/multiproject_samplebuild/multiproject/webRoot/slowfei.com			// web static file access directory
	//		$GOPATH/multiproject_samplebuild/multiproject/webRoot/blog.slowfei.com		// web static file access directory
	//		$GOPATH/multiproject_samplebuild/src										// source files
	//		$GOPATH/multiproject_samplebuild/main.go									// main file
	//
	//		www.slowfei.com use slowfei.com
	//		注意：添加("localhost") 主要是为了方便演示，进入引导的静态页面。
	server.SetMultiProjectHosts("slowfei.com", "blog.slowfei.com", "localhost")
	//	手动设置 tls 配置参数
	server.SetHttpTLS("multiproject/config/cert.pem", "multiproject/config/key.pem", 8081, false)

	/*
		slowfei.com 路由操作
		controller 需要设置相应的 host 否则无法访问
		SetScheme 设置路由的访问scheme，默认是同时支持 http || https
	*/
	wwwOption := router.DefaultReflectRouterOption().SetHost("slowfei.com").SetScheme(leafveingo.URI_SCHEME_HTTP)
	server.AddRouter(router.CreateReflectControllerWithOption("/", www.WWWController{}, wwwOption))

	//	https 访问登录页面
	//	注意这里是使用 RESTful router 控制器
	wwwLoginOption := router.DefaultRESTfulRouterOption().SetHost("slowfei.com").SetScheme(leafveingo.URI_SCHEME_HTTPS)
	server.AddRouter(router.CreateRESTfulControllerWithOption("/login", www.WWWLoginController{}, wwwLoginOption))

	/*
		blog.slowfei.com 路由操作
	*/
	blogOption := router.DefaultReflectRouterOption().SetHost("blog.slowfei.com")
	server.AddRouter(router.CreateReflectControllerWithOption("/", blog.BlogController{}, blogOption))

	/*
		-----------------
		admin server
		将后台管理的服务用另一个端口进行访问
		SetPort(0)默认会配置8080。
		如果需要只开启https SetPort(0)可以随意设置，但是SetHttpTLS(...,...,true) aloneRun必须设置独立运行
	*/
	optionAdmin := leafveingo.DefaultOption().SetConfigPath("admin/config/app.conf").SetAddr("127.0.0.1").SetPort(0).SetSMGCTime(300)
	serverAdmin := leafveingo.NewLeafveinServer("admin", optionAdmin)
	serverAdmin.SetHttpTLS("multiproject/config/cert.pem", "multiproject/config/key.pem", 8090, true)

	// 后台支持https访问
	adminRouterOption := router.DefaultReflectRouterOption().SetScheme(leafveingo.URI_SCHEME_HTTPS)
	serverAdmin.AddRouter(router.CreateReflectControllerWithOption("/", &admin.AdminController{}, adminRouterOption))

	leafveingo.Start()
}
