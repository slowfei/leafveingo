package controller

import (
	"github.com/slowfei/leafveingo"
)

//	模板演示控制器
type TemplateController struct {
	tag string
}

// default index
func (t *TemplateController) Index() interface{} {
	params := make(map[string]interface{})
	params["Content"] = "Hello Template Index"

	// 模版默认路径
	// lv_reflect_router.go 实现规则是[router key]/[func name].tpl
	//		router key = "/admin/"
	//			   URL = POST http://localhost:8080/admin/login
	//		 func name = PostLogin
	//	 template path = template/admin/PostLogin.tpl

	// lv_restful_router.go	实现规则是[router key]/[func name].tpl
	//		router key = "/api/object"
	//			   URL = DELETE http://localhost:8080/api/object
	//		 func name = delete
	//	 template path = template/api/object/delete.tpl

	//	当前请求URL http://localhost:8080/r/template
	//    router key  = "/t/"
	//	template path = template/r/template.tpl
	//	对应的模板位置 github.com/slowfei/leafveingo/example/sample/SampleWeb/template/t/index.tpl
	return leafveingo.BodyTemplate(params)

	//	另外一种模板加载方式
	//	指定模板路径进行加载，模板指定的路径是相对路径，
	//	以设置的leafvein.SerTemplateDir("template")开始进行查找
	// return leafveingo.BodyTemplateByTplPath("/t/index.tpl", data)
}
