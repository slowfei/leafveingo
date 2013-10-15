package controller

import (
	"fmt"
	"testing"
)

func TestRouterMethodParse(t *testing.T) {
	arc := RouterController{}

	requrl := "/router/forum-10-20"
	fmt.Println("url = ", requrl)
	m, p := arc.RouterMethodParse(requrl)
	fmt.Println("methodName:", m)
	fmt.Println("params:", p)
	fmt.Println("")

	requrl = "/router/thread-10-20-30"
	fmt.Println("url = ", requrl)
	m, p = arc.RouterMethodParse(requrl)
	fmt.Println("methodName:", m)
	fmt.Println("params:", p)
	fmt.Println("")

	requrl = "/router/space/uid/12/"
	fmt.Println("url = ", requrl)
	m, p = arc.RouterMethodParse(requrl)
	fmt.Println("methodName:", m)
	fmt.Println("params:", p)
	fmt.Println("")

	requrl = "/Router/C4CA4238A0B923820DCC509A6F75849B"
	fmt.Println("url = ", requrl)
	m, p = arc.RouterMethodParse(requrl)
	fmt.Println("methodName:", m)
	fmt.Println("params:", p)
	fmt.Println("")

}
