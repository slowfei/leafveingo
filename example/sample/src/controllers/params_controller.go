package controller

import (
	"bytes"
	"fmt"
	"github.com/slowfei/leafveingo"
	"github.com/slowfei/leafveingo/session"
	"net/http"
)

//	演示控制器参数
type ParamsController struct {
	tag string
}

//	index
func (p *ParamsController) Index() string {
	return "sorry not link address..."
}

//	获取 *http.Request 和 http.ResponseWriter参数
func (p *ParamsController) Param1(request *http.Request, rw http.ResponseWriter) string {
	return fmt.Sprintf("Request:%v \n\n ResponseWriter:%p", request, rw)
}

//	获取session
func (p *ParamsController) Param2(session LVSession.HttpSession) string {
	return fmt.Sprintf("HttpSession:%p", session)
}

//	获取上下文HttpContext
func (p *ParamsController) Param3(context *leafveingo.HttpContext) string {
	return fmt.Sprintf("HttpContext:%p", context)
}

//	参数结构的封装
func (p *ParamsController) Param4(params struct {
	Id   string // id 切记首字母大写，其他与参数匹配 Id = id, ID = ID | iD
	Name string // name
}) string {
	return fmt.Sprintf("id=%v, name=%v", params.Id, params.Name)
}

type UserInfo struct {
	Id   string
	Name string
}

func (p *ParamsController) Param5(params UserInfo) string {
	return fmt.Sprintf("id=%v, name=%v", params.Id, params.Name)
}

/******* 嵌套参数结构封装 *******/
type UserType struct {
	TypeName string //	用户的类型名称
}
type User struct {
	Uid      int      // 用户ID
	UserName string   // 用户名称
	Type     UserType // 用户类型
	Interest []string // 用户的兴趣爱好，集合参数
}

//	页面请求
func (p *ParamsController) Param6() interface{} {
	//	参数名设置注意查看input的name
	bodyHTML := `
<!doctype html>
<html>
	<meta charset="UTF-8">
	<head>
		<title>嵌套参数结构封装</title>
	</head>
	<body>
		<form action="param6.htm" method="post">
			<lable>用户ID</label>
			<input type="text" name="uid" value="1"/>
			<br/>

			<lable>用户名称</label>
			<input type="text" name="userName" value="slowfei"/>
			<br/>

			<lable>用户类型（name="type.typeName" 对应结构体的字段名称）</label>
			<input type="text" name="type.typeName" value="admin"/>
			<br/>

			<input type="submit" value="提交" />
		</form>
	</body>
</html>
	`
	// return leafveingo.Bodybyte([]byte(bodyHTML), true, "text/html; charset=utf-8", nil)
	return leafveingo.BodyHtml(bodyHTML)
}

//	post 请求打印参数
func (p *ParamsController) PostParam6(params User) string {
	return fmt.Sprintf("uid=%v\nuserName=%v\ntype=%v", params.Uid, params.UserName, params.Type.TypeName)
}

/******* 数组集合参数封装 ********/
type UserBean struct {
	Users []User
	Tags  []string
}

//	页面请求
func (p *ParamsController) Param7() leafveingo.HtmlOut {
	//	参数名设置注意查看input的name
	bodyHTML := `
<!doctype html>
<html>
	<meta charset="UTF-8">
	<head>
		<title>嵌套参数结构封装</title>
	</head>
	<body>
		<form action="param7.htm" method="post">
			<h3>用户1</h3>
			<lable>用户ID(name="users[0].uid"): 用户名称(name="users[0].userName"): 用户类型(name="users[0].type.typeName")</label>
			<br/>
			<input type="text" name="users[0].uid" value="1"/>
			<input type="text" name="users[0].userName" value="slowfei_1"/>
			<input type="text" name="users[0].type.typeName" value="admin"/>
			<br/>
			<label>兴趣爱好(name="users[0].interest[0]")</label>
			<br/>
			<input type="text" name="users[0].interest[0]" value="1_爱好1"/>
			<input type="text" name="users[0].interest[1]" value="1_爱好2"/>
			<input type="text" name="users[0].interest[2]" value="1_爱好3"/>
			<br/>

			<h3>用户2</h3>
			<lable>用户ID(name="users[1].uid"): 用户名称(name="users[1].userName"): 用户类型(name="users[1].type.typeName")</label>
			<br/>
			<input type="text" name="users[1].uid" value="2"/>
			<input type="text" name="users[1].userName" value="slowfei_2"/>
			<input type="text" name="users[1].type.typeName" value="admin"/>
			<br/>
			<label>兴趣爱好(name="users[1].interest[0]")</label>
			<br/>
			<input type="text" name="users[1].interest[0]" value="2_爱好1"/>
			<input type="text" name="users[1].interest[1]" value="2_爱好2"/>
			<input type="text" name="users[1].interest[2]" value="2_爱好3"/>
			<br/>
			
			<h3>用户3</h3>
			<lable>用户ID(name="users[2].uid"): 用户名称(name="users[2].userName"): 用户类型(name="users[2].type.typeName")</label>
			<br/>
			<input type="text" name="users[2].uid" value="3"/>
			<input type="text" name="users[2].userName" value="slowfei_3"/>
			<input type="text" name="users[2].type.typeName" value="admin"/>
			<br/>
			<label>兴趣爱好(name="users[2].interest[0]")</label>
			<br/>
			<input type="text" name="users[2].interest[0]" value="3_爱好1"/>
			<input type="text" name="users[2].interest[1]" value="3_爱好2"/>
			<input type="text" name="users[2].interest[2]" value="3_爱好3"/>
			<br/>

			<h3>Tags</h3>
			tags1(name="tags"):
			<input type="checkbox" name="tags" value="tags1">
			tags2(name="tags"):
			<input type="checkbox" name="tags" value="tags2">
			tags3(name="tags"):
			<input type="checkbox" name="tags" value="tags3">
			<br/>

			<input type="submit" value="提交" />
		</form>
	</body>
</html>
	`

	return leafveingo.BodyHtml(bodyHTML)
}

//	post 打印集合参数
func (p *ParamsController) PostParam7(params UserBean) string {
	buf := bytes.NewBufferString("")

	buf.WriteString(fmt.Sprintf("Tags: %v \n\n", params.Tags))

	buf.WriteString(fmt.Sprintf("用户数量:%v \n", len(params.Users)))
	for i, v := range params.Users {
		buf.WriteString(fmt.Sprintf("用户%v: %v \n", i, v))
	}

	return buf.String()
}
