package LVRouter

import (
	"github.com/slowfei/gosfcore/log"
	lv "github.com/slowfei/leafveingo"
)

var (
	LogConfigNone = `
		{
			"InitAppenders":[
				"console"
			],
			"LogGroups" :{
				"globalGroup" :{
					"Appender":[
						"console"
					],
					"none":true,
					"ConsolePattern":"${yyyy}-${MM}-${dd} ${hh}:${mm}:${ss}${SSSSSS} [${TARGET}] ([${LOG_GROUP}][${LOG_TAG}][L${FILE_LINE} ${FUNC_NAME}])\n${MSG}"
				}
			}
		}
	`

	Template_HTML = `
<!doctype html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<title>Template</title>
</head>
<body>
	<h1>Hello Template {{.}}.</h1>
</body>
</html>
	`

	Server = lv.NewLeafveinServer("TestRouter", lv.DefaultOption())
)

func init() {
	// runtime.GOMAXPROCS(runtime.NumCPU())
	SFLog.StartLogManager(3000)
	SFLog.LoadConfigByJson([]byte(LogConfigNone))

	Server.AddRouter(CreateReflectController("/temp/", TestReflectController{}))
	Server.AddRouter(CreateReflectController("/temp2/", TestReflectController{}))
	Server.AddRouter(CreateReflectController("/temp3/", TestReflectController{}))
	Server.AddRouter(CreateReflectController("/placeholder/", TestReflectController{}))
	Server.AddRouter(CreateReflectController("/placeholder2/", TestReflectController{}))
	Server.AddRouter(CreateReflectController("/placeholder3/", TestReflectController{}))
	Server.AddRouter(CreateReflectController("/url/", TestReflectController{}))
	Server.AddRouter(CreateReflectController("/url2/", TestReflectController{}))
	Server.AddRouter(CreateReflectController("/url3/", TestReflectController{}))
	Server.AddRouter(CreateReflectController("/url4/", TestReflectController{}))

	Server.AddRouter(CreateReflectController("/", TestReflectController{}))
	Server.AddRouter(CreateRESTfulController("/restful/", TestRESTfulRouterController{}))

	Server.AddCacheTemplate("template", Template_HTML)

	Server.TestStart()
}
