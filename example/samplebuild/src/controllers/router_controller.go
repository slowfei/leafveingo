package controller

import (
	"fmt"
	"github.com/slowfei/leafveingo"
	"regexp"
)

var (
	_urlrexForum  = regexp.MustCompile(`forum-([0-9]+)-([0-9]+)(\.\w+)?`)
	_urlrexThread = regexp.MustCompile(`thread-(?P<bid>[0-9]+)-(?P<tid>[0-9]+)-(?P<pid>[0-9]+)`)
	_urlrexSpace  = regexp.MustCompile(`space/(username|uid)/(.+)/`)
	_urlrexMD5    = regexp.MustCompile(`[0-9a-zA-Z]{32}`)
)

//	高级路由器演示控制器
type RouterController struct {
	tag string
}

//	解析的URL
//	http://localhost:8080/router/forum-([0-9]+)-([0-9]+)(\.\w+)?
//	http://localhost:8080/router/thread-([0-9]+)-([0-9]+)-([0-9]+)
//	http://localhost:8080/router/space/(username|uid)/(.+)/
//	http://localhost:8080/router/[0-9a-zA-Z]{32}
func (arc RouterController) RouterMethodParse(option *leafveingo.RouterOption) (funcName string, params map[string]string) {
	//	下面演示几种正则的解析操作，聪明的您肯定会以最优的处理方式来返回所需要的函数名和参数。
	requrl := option.RouterPath
	switch {
	case _urlrexForum.MatchString(requrl):
		ps := _urlrexForum.FindStringSubmatch(requrl)
		if 3 <= len(ps) {
			params = map[string]string{"bid": ps[1], "tid": ps[2]}
			funcName = "Forum"
			return
		}
	case _urlrexThread.MatchString(requrl):
		names := _urlrexThread.SubexpNames()
		ps := _urlrexThread.FindStringSubmatch(requrl)

		nCount := len(names)
		pCount := len(ps)

		if nCount == pCount && 1 < nCount {
			params = make(map[string]string)
			for i := 1; i < pCount; i++ {
				params[names[i]] = ps[i]
			}
			funcName = "Thread"
			return
		}

	case _urlrexSpace.MatchString(requrl):
		ps := _urlrexSpace.FindStringSubmatch(requrl)
		if 3 == len(ps) {
			params = map[string]string{ps[1]: ps[2]}
			funcName = "Space"
			return
		}
	case _urlrexMD5.MatchString(requrl):
		params = map[string]string{"md5": requrl}
		funcName = "MD5"
		return
	default:
		fmt.Println("defalut")
	}

	return
}

func (arc *RouterController) Forum(params struct {
	Bid int
	Tid int
}) string {
	return fmt.Sprintf("Bid=%v\nTid=%v", params.Bid, params.Tid)
}

//	url 后缀为json
func (arc *RouterController) ForumJson(params struct {
	Bid int
	Tid int
}) interface{} {
	return leafveingo.BodyJson(params)
}

func (arc *RouterController) Thread(params struct {
	Bid int
	Tid int
	Pid int
}) string {
	return fmt.Sprintf("Bid=%v\nTid=%v\nPid=%v", params.Bid, params.Tid, params.Pid)
}

func (arc *RouterController) PostThread(params struct {
	Bid int
	Tid int
	Pid int
}) string {
	return fmt.Sprintf("Bid=%v\nTid=%v\nPid=%v", params.Bid, params.Tid, params.Pid)
}

func (arc *RouterController) Space(params struct {
	Uid      int
	Username string
}) string {
	return fmt.Sprintf("Uid=%v\nUsername=%v", params.Uid, params.Username)
}

func (arc *RouterController) MD5(params struct {
	Md5 string
}) string {
	return params.Md5
}
