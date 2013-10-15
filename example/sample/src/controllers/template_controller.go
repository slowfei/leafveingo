package controller

import (
	"github.com/slowfei/leafveingo"
)

//	模板演示控制器
type TemplateController struct {
	tag string
}

// index
func (t *TemplateController) Index() interface{} {
	params := make(map[string]interface{})
	params["Content"] = "Hello Template Index"
	//	对应的模板位置 github.com/slowfei/leafveingo/example/sample/SampleWeb/template/t/index.tpl
	return leafveingo.BodyTemplate(params)
}
