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
	//	对应的模板位置 github.com/slowfei/leafveingo/example/sample/SampleWeb/template/t/index.tpl
	//	router key  = "/t/"
	//	request url = http://localhost:8080/t/
	return leafveingo.BodyTemplate(params)

	//	另外一种模板加载方式
	//	指定模板路径进行加载，模板指定的路径是相对路径，
	//	以设置的leafvein.SerTemplateDir("template")开始进行查找
	// return leafveingo.BodyTemplateByTplPath("/t/index.tpl", data)
}
