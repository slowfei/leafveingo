package controller

import (
	"fmt"
	"github.com/slowfei/gosfcore/encoding/json"
	"github.com/slowfei/leafveingo"
)

//	控制器返回参数演示
type ReturnParamController struct {
	tag string
}

//	输出 text, Content-Type = text/plain
func (r *ReturnParamController) Text() string {
	// return "return text"
	return leafveingo.BodyText("return text")
}

// 输出text html, Content-Type = text/plain
func (r *ReturnParamController) Html() leafveingo.HtmlOut {
	html := `
		<!doctype html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<title>Document</title>
		</head>
		<body>
			<h1>Hello world</h1>
		</body>
		</html>
	`
	return leafveingo.BodyHtml(html)
}

//	输出json Content-Type = application/json
func (r *ReturnParamController) Json(params struct {
	T int
}) interface{} {
	type TStruct struct {
		ID   string
		Name string
	}
	t := TStruct{"1", "slowfei"}
	t2 := TStruct{"2", "slowfei_2"}

	var j SFJson.Json

	if params.T == 1 {
		j = leafveingo.BodyJson([]TStruct{t, t2})
	} else {
		j = leafveingo.BodyJson(t)
	}

	return j
}

//	输出 []byte
func (r *ReturnParamController) Byte() leafveingo.ByteOut {
	return leafveingo.BodyByte([]byte("hello world []byte"), "text/plain; charset=utf-8", nil)
}

//	输出模板
func (r *ReturnParamController) Template(params struct {
	Info string
	T    int
}) interface{} {
	if params.T == 1 {
		return leafveingo.BodyTemplateByTplPath("custom/custom.tpl", params.Info)
	} else {
		// 模板默认路径，根据router key 命名文件夹然后函数名指定模板名称
		// github.com/slowfei/leafveingo/example/sample/SampleWeb/template/r/template.tpl
		return leafveingo.BodyTemplate(params.Info)
	}

}

//	重定向url
func (r *ReturnParamController) Redirect(params struct {
	Url string
}) interface{} {
	if 0 == len(params.Url) {
		return leafveingo.BodyHtml(`
			<!doctype html>
			<html lang="en">
			<head>
				<meta charset="UTF-8">
				<title>Redirect</title>
			</head>
			<body>
				<a href="/r/redirect.htm?url=http://www.google.com">Redirect to google</a>
			</body>
			</html>
			`)
	} else {
		return leafveingo.BodyRedirect(params.Url)
	}
}

//	控制器转发
func (r *ReturnParamController) Dispatcher(params struct {
	Info string
}) interface{} {
	//	注意router key，对应添加控制器router
	return leafveingo.BodyCellController("/r/", "DispTest")
}
func (r *ReturnParamController) DispTest(params struct {
	Info string
}) string {
	return fmt.Sprintf("Dispatcher to DispTest Info = %v", params.Info)
}

//	文件输出
func (r *ReturnParamController) File() interface{} {
	//	webRoot目录为标准
	//	github.com/slowfei/leafveingo/example/sample/SampleWeb/webRoot/file.zip
	return leafveingo.BodyServeFile("file.zip")
}
