package leafveingo

import (
	"fmt"
	"net/url"
	"reflect"
	"testing"
	"time"
)

type Hobby struct {
	Name  string
	Names []string
}

type UserType struct {
	TypeName string
}

type User struct {
	UserName string
	UserPwd  string
	Type     UserType
	TypeP    *UserType
	Hobbys   []Hobby
	HobbysP  []*Hobby
	Content  []byte
}

type Params struct {
	Users    []User
	UsersP   []*User
	OneUser  User
	OneUserP *User
}

//	测试单个设值测试，此函数主要用户调试测试
func TestStringNameSetStructFieldValue(t *testing.T) {
	leafvein := &sfLeafvein{}

	keys := url.Values{}
	// keys.Add("users[0].hobbys[0].name", "hobbys0-0")
	// keys.Add("users[0].hobbys[1].name", "hobbys0-1")
	// keys.Add("users[0].hobbys[2].name", "hobbys0-2")
	// keys.Add("users[1].hobbys[0].name", "hobbys1-0")
	// keys.Add("users[1].hobbys[1].name", "hobbys1-1")
	// keys.Add("users[0].typeP.typeName", "0-type-1")
	// keys.Add("users[0].type.typeName", "0-type-2")
	// keys.Add("users[1].hobbysP[0].name", "hobbysP1-0")
	// keys.Add("users[1].hobbysP[1].name", "hobbysP1-1")
	// keys.Add("oneUser.userName", "oneName")
	// keys.Add("oneUser.userPwd", "onePwd")
	// keys.Add("oneUser.hobbys[0].names[0]", "one-0-bobys-0-names-0")
	// keys.Add("oneUser.hobbys[1].names[0]", "one-0-bobys-1-names-0")
	keys.Add("usersP[0].hobbys[0].name", "userP0-hobbys0-0")
	keys.Add("usersP[0].hobbys[1].name", "userP0-hobbys0-1")
	keys.Add("usersP[1].hobbys[0].name", "userP1-hobbys1-0")

	refValue := leafvein.newStructPtr(reflect.TypeOf(Params{}), keys)

	// str := "userP0-hobbys0-0"
	// leafvein.setStructFieldValue(refValue, "usersP[0].hobbys[0].name", &str)

	//	转换类型然后进行比较判断
	params := refValue.Interface().(*Params)

	fmt.Println(len(params.UsersP[0].Hobbys))
}

//	设置结构字段测试
func funcSetStructFieldValue() *Params {
	leafvein := &sfLeafvein{}

	keys := url.Values{}
	keys.Add("users[0].hobbys[0].name", "hobbys0-0")
	keys.Add("users[0].hobbys[1].name", "hobbys0-1")
	keys.Add("users[0].hobbys[2].name", "hobbys0-2")
	keys.Add("users[1].hobbys[0].name", "hobbys1-0")
	keys.Add("users[1].hobbys[1].name", "hobbys1-1")
	keys.Add("users[0].typeP.typeName", "0-type-1")
	keys.Add("users[0].type.typeName", "0-type-2")
	keys.Add("users[1].hobbysP[0].name", "hobbys1-0")
	keys.Add("users[1].hobbysP[1].name", "hobbys1-1")
	keys.Add("oneUser.userName", "oneName")
	keys.Add("oneUser.userPwd", "onePwd")

	start := time.Now()
	refValue := leafvein.newStructPtr(reflect.TypeOf(Params{}), keys)
	fmt.Println("newStructPtr time :", time.Now().Sub(start))
	start = time.Now()
	for k, v := range keys {
		leafvein.setStructFieldValue(refValue, k, v[0])
	}
	fmt.Println("setStructFieldValue time :", time.Now().Sub(start))
	return refValue.Interface().(*Params)
}

//	测试设置结构字段直
func TestSetStructFieldValue(t *testing.T) {
	start := time.Now()
	params := funcSetStructFieldValue()
	fmt.Println("time :", time.Now().Sub(start))

	if nil == params.OneUserP {
		t.Error("nil != params.OneUserP 初始化错误")
		return
	}

	if 2 != len(params.Users) {
		t.Error("2 != len(params.Users) 初始化的用户数量不正确")
		return
	}

	if 3 != len(params.Users[0].Hobbys) {
		t.Error("3 != len(params.Users[0].Hobbys) 初始化数量不正确")
		return
	}

	if 2 != len(params.Users[1].HobbysP) {
		t.Error("2 != len(params.Users[1].HobbysP 初始化数量不正确")
		return
	}

	if "oneName" != params.OneUser.UserName {
		t.Error("oneName != params.OneUser.UserName 设值不正确")
		return
	}

	if "onePwd" != params.OneUser.UserPwd {
		t.Error("onePwd != params.OneUser.UserPwd 设值不正确")
		return
	}

	if "hobbys0-0" != params.Users[0].Hobbys[0].Name {
		t.Error(`"hobbys0-0" != params.Users[0].Hobbys[0].Name`)
		return
	}

	if "hobbys0-1" != params.Users[0].Hobbys[1].Name {
		t.Error(`"hobbys0-1" != params.Users[0].Hobbys[1].Name`)
		return
	}

	if "hobbys0-2" != params.Users[0].Hobbys[2].Name {
		t.Error(`"hobbys0-2" != params.Users[0].Hobbys[2].Name`)
		return
	}

	if "0-type-1" != params.Users[0].TypeP.TypeName {
		t.Error(`""0-type-1" != params.Users[0].TypeP.TypeName`)
		return
	}

	if "0-type-2" != params.Users[0].Type.TypeName {
		t.Error(`""0-type-2" != params.Users[0].Type.TypeName`)
		return
	}

	if "hobbys1-0" != params.Users[1].Hobbys[0].Name {
		t.Error(`"hobbys1-0" != params.Users[1].Hobbys[0].Name`)
		return
	}

	if "hobbys1-1" != params.Users[1].Hobbys[1].Name {
		t.Error(`"hobbys1-1" != params.Users[1].Hobbys[1].Name`)
		return
	}
}

//	提供TestNewStructPtr和Benchmark_NewStructPtr进行测试
func funcNewStructPtr() interface{} {

	leafvein := &sfLeafvein{}

	keys := url.Values{}
	keys.Add("users[0].hobbys[0].name", "hobbys1")
	keys.Add("users[0].hobbys[1].name", "hobbys1")
	keys.Add("users[0].hobbys[2].name", "hobbys1")
	keys.Add("users[1].hobbysP[0].name", "hobbys1")
	keys.Add("users[1].hobbysP[1].name", "hobbys1")
	keys.Add("usersP[0].hobbys[0].name", "hobbys1")
	keys.Add("oneUser.userName", "oneName")
	keys.Add("oneUser.userPwd", "onePwd")
	keys.Add("oneUser.hobbys[0].names[0]", "one-0-hobbys-0-names-0")
	keys.Add("oneUser.hobbys[0].names[1]", "one-0-hobbys-0-names-1")
	keys.Add("oneUser.hobbys[0].names[2]", "one-0-hobbys-0-names-2")
	keys.Add("oneUser.hobbys[1].names[0]", "one-0-hobbys-1-names-1")

	return leafvein.newStructPtr(reflect.TypeOf(Params{}), keys).Interface()
}

//	测试创建结构里的嵌套内容
func TestNewStructPtr(t *testing.T) {
	start := time.Now()

	params := funcNewStructPtr().(*Params)

	//	TODO 测试速度有点担忧670.976us 最好控制在100us以内
	fmt.Println("time:", time.Now().Sub(start))

	if nil == params.OneUserP {
		t.Error("nil != params.OneUserP 初始化错误")
		return
	}

	if 2 != len(params.Users) {
		t.Error("2 != len(params.Users) 初始化的用户数量不正确")
		return
	}

	if 3 != len(params.Users[0].Hobbys) {
		t.Error("3 != len(params.Users[0].Hobbys) 初始化数量不正确")
		return
	}

	if 2 != len(params.Users[1].HobbysP) {
		t.Error("2 != len(params.Users[1].HobbysP 初始化数量不正确")
		return
	}

	if 2 != len(params.OneUser.Hobbys) {
		t.Error("2 != len(params.OneUser.Hobbys) 初始化数量不正确")
		return
	}

	if 3 != len(params.OneUser.Hobbys[0].Names) {
		t.Error("3 != len(params.OneUser.Hobbys[0].Names) 初始化数量不正确")
		return
	}

	if 1 != len(params.OneUser.Hobbys[1].Names) {
		t.Error("1 != len(params.OneUser.Hobbys[1].Names) 初始化数量不正确")
		return
	}
}

//	速度测试new struct, 包含里面的指针结构类型和切片大小分配
func Benchmark_NewStructPtr(b *testing.B) {
	for i := 0; i < b.N; i++ {
		params := funcNewStructPtr().(*Params)

		if nil == params.OneUserP {
			b.Error("nil != params.OneUserP 初始化错误")
			return
		}

		if 2 != len(params.Users) {
			b.Error("2 != len(params.Users) 初始化的用户数量不正确")
			return
		}

		if 3 != len(params.Users[0].Hobbys) {
			b.Error("3 != len(params.Users[0].Hobbys) 初始化数量不正确")
			return
		}

		if 2 != len(params.Users[1].HobbysP) {
			b.Error("2 != len(params.Users[1].HobbysP 初始化数量不正确")
			return
		}

		if 2 != len(params.OneUser.Hobbys) {
			b.Error("2 != len(params.OneUser.Hobbys) 初始化数量不正确")
			return
		}

		if 3 != len(params.OneUser.Hobbys[0].Names) {
			b.Error("3 != len(params.OneUser.Hobbys[0].Names) 初始化数量不正确")
			return
		}

		if 1 != len(params.OneUser.Hobbys[1].Names) {
			b.Error("1 != len(params.OneUser.Hobbys[1].Names) 初始化数量不正确")
			return
		}

	}
}
