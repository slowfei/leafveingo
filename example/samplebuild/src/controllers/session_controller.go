package controller

import (
	"fmt"
	"github.com/slowfei/leafveingo"
	//	原生的package是LVSession，这只是slowfei的习惯将相同的包进行区分，如果不习惯的您可以修改自己需要设置的包名。
	session "github.com/slowfei/leafveingo/session"
)

//	http session 演示控制器
type SessionController struct {
	tag string
}

//	基本的获取session
func (s *SessionController) Session(sess session.HttpSession) string {

	outString := ""

	name, ok := sess.Get("name")
	if !ok {
		//	测试存储数据
		sess.Set("name", "slowfei")
		outString = "Has not been set to the name, refreshed look."
	} else {
		outString = fmt.Sprintf("%v, welcome", name.(string))
	}

	return fmt.Sprintf("get the session address: %p\n%v", sess, outString)
}

//	上下文对session的操作
func (s *SessionController) Context(context *leafveingo.HttpContext) string {
	//	cookie token会在一定时间会自动重置，这里就不选择重置了。
	//	看需求而定是否需要重置。
	sess, err := context.Session(false)

	if nil != err {
		return fmt.Sprintf("session get error:%v", err.Error())
	}

	outString := ""

	name, ok := sess.Get("name")
	if !ok {
		//	测试存储数据
		sess.Set("name", "slowfei")
		outString = "Has not been set to the name, refreshed look."
	} else {
		outString = fmt.Sprintf("%v, welcome", name.(string))
	}

	return fmt.Sprintf("context get the session address: %p\n%v", sess, outString)
}

//	演示form token的操作
func (s *SessionController) Form(context *leafveingo.HttpContext) leafveingo.HtmlOut {
	token := context.FormTokenHTML()

	html := `
<!doctype html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<title>Form Token</title>
</head>
<body>
	尝试提交验证form token
	<br/>
	<form action="form.htm" method="post">
		` + token + `
	<input type="submit" value="提交">
	</form>
</body>
</html>
	`
	return leafveingo.BodyHtml(html)
}

//	演示form token的验证操作
func (s *SessionController) PostForm(context *leafveingo.HttpContext) string {
	if context.CheckFormToken() {
		return "form token check success."
	} else {
		return "form token check fail."
	}
}
