package controller

import (
	. "github.com/slowfei/leafveingo"
)

//	状态页 演示控制器
type StatusController struct {
	tag string
}

//	500错误
func (s *StatusController) S500() {
	panic("500 error...")
}

//	自定义403模版
func (s *StatusController) S403() HttpStatusValue {
	//	github.com/slowfei/leafveingo/example/samplebuild/sample/template/403.tpl
	//	在模版的根目录建立相应代码状态的模版例如：("403.tpl")
	return BodyStatusPage(Status403, "无权访问", "", "")
}
