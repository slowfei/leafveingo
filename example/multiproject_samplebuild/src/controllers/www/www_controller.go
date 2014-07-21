package www

import (
	lv "github.com/slowfei/leafveingo"
)

//
//	slowfei.com controller
//
type WWWController struct {
	tag string
}

func (w *WWWController) Index() string {
	return "Hello world, slowfei.com"
}

//
//	login congroller
//
type WWWLoginController struct {
	tag string
}

func (l *WWWLoginController) Get(context *lv.HttpContext) interface{} {
	return "slowfei.com https login access"
}
func (l *WWWLoginController) Post(context *lv.HttpContext) interface{} {
	return lv.BodyStatusPage(lv.Status404, "404", "", "")
}
func (l *WWWLoginController) Put(context *lv.HttpContext) interface{} {
	return lv.BodyStatusPage(lv.Status404, "404", "", "")
}
func (l *WWWLoginController) Delete(context *lv.HttpContext) interface{} {
	return lv.BodyStatusPage(lv.Status404, "404", "", "")
}
func (l *WWWLoginController) Header(context *lv.HttpContext) interface{} {
	return lv.BodyStatusPage(lv.Status404, "404", "", "")
}
func (l *WWWLoginController) Options(context *lv.HttpContext) interface{} {
	return lv.BodyStatusPage(lv.Status404, "404", "", "")
}
func (l *WWWLoginController) Other(context *lv.HttpContext) interface{} {
	return lv.BodyStatusPage(lv.Status404, "404", "", "")
}
