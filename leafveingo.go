//	Copyright 2013 slowfei And The Contributors All rights reserved.
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.
//
//  Create on 2013-8-16
//  Update on 2013-10-23
//  Email  slowfei@foxmail.com
//  Home   http://www.slowfei.com
//	version 0.0.1.000

//	web框架leafveingo
package leafveingo

import (
	"errors"
	"flag"
	"fmt"
	"github.com/slowfei/gosfcore/helper"
	"github.com/slowfei/gosfcore/log"
	"github.com/slowfei/gosfcore/utils/filemanager"
	"github.com/slowfei/gosfcore/utils/strings"
	"github.com/slowfei/leafveingo/session"
	"github.com/slowfei/leafveingo/template"
	"net"
	"net/http"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"
)

const (

	//	leafveingo version
	VERSION string = "0.0.1.000"

	//	controller default url request ("http://localhost:8080/") method
	CONTROLLER_DEFAULT_METHOD = "Index"
	//	请求访问到控制器前进行调用的函数名称
	//	Requesting access to the controller before calling the function name
	CONTROLLER_BEFORE_METHOD = "Before"
	//	After the calling the function name
	CONTROLLER_AFTER_METHOD = "After"

	// template func key
	TEMPLATE_FUNC_KEY_VERSION     = "Leafveingo_version"
	TEMPLATE_FUNC_KEY_APP_VERSION = "Leafveingo_app_version"
	TEMPLATE_FUNC_KEY_APP_NAME    = "Leafveingo_app_name"

	//	不知道为什么golang使用http.ServeFile判断为此结尾就重定向到/index.html中。
	//	这里定义主要用来解决这个问题的处理
	INDEX_PAGE = "/index.html"

	//	http status support code
	//	http://zh.wikipedia.org/wiki/HTTP%E7%8A%B6%E6%80%81%E7%A0%81
	HTTP_STATUS_CODE_200 = 200
	HTTP_STATUS_CODE_301 = 301
	HTTP_STATUS_CODE_307 = 307
	HTTP_STATUS_CODE_400 = 400
	HTTP_STATUS_CODE_404 = 404
	HTTP_STATUS_CODE_403 = 403
	HTTP_STATUS_CODE_500 = 500
	HTTP_STATUS_CODE_503 = 503
)

var (
	//	一个变量的存储区域，用于存储已经初始化好的Leafvein
	_thisLeafvein ISFLeafvein
	//	开发模式命令
	_flagDeveloper bool
	//
	_rexW = regexp.MustCompile("\\w+")

	lvLog = SFLog.NewLogger("Leafveingo")
)

func init() {
	flag.BoolVar(&_flagDeveloper, "devel", false, "developer mode start.")

}

//	ISFLeafvein main interface
type ISFLeafvein interface {
	//	start leafveingo
	Start()

	//	close leafveingo
	Close()
	//	leafveion version info
	Version() string

	//	获取leafveingo的一个Handler, 方便在其他应用进行扩展leafveingo
	//	get leafveingo handler, Easily be extended to other applications
	//	@prefix	 url prefix
	GetHandlerFunc(prefix string) (handler http.Handler, err error)

	//	add controller
	//
	//	@routerKey 	需要访问url路由设置。 e.g.: / == http://127.0.0.1/; admin/ == http://127.0.0.1/admin/
	//	@controller	控制器分为地址传递和值传递
	//				值传递：
	//				AddControllers("/Home/", HomeController{})
	//				每次请求(http://127.0.0.1/home/)时 controller 都会根据设置的控制器类型新建立一个对象进行处理
	//
	//				地址传递：
	//				AddControllers("/Admin/", &AdminController{})
	//				跟值传递相反，每次请求时都会使用设置的控制器地址进行处理，应用结束也不会改变，每次请求控制器都不会改变内存地址
	//				这里涉及到并发时同时使用一个内存地址处理的问题，不过目前还没有弄到锁，并发后期leafveingo会进行改进和处理。
	AddController(routerKey string, controller interface{})

	//	web application map
	Application() SFHelper.Map

	//	http session manager
	HttpSessionManager() *LVSession.HttpSessionManager
	//	is use session, default true
	SetUseSession(b bool)
	//	set is auto GC session clean work, default true
	SetGCSession(b bool)

	//	set http session maxlife time, unit second
	//	default 1800 second(30 minute)
	//	must be greater than 60 seconds
	SetSessionMaxlifeTime(second int32)
	SessionMaxlifeTime() int32

	//	developer mode
	IsDevel() bool
	// is starting
	IsStart() bool

	//	current leafveingo the file operation directory
	//	operatingDir 是根据 appName 建立的目录路径
	OperatingDir() string

	//	web directory, can read and write to the directory
	//	primary storage html, css, js, image, zip the file
	//	OperatingDir() created under the directory
	//	default webRoot
	//	@name "webRoot" "web"...
	SetWebRootDir(name string)
	WebRootDir() string

	//	template directory, primary storage template file
	//	OperatingDir() created under the directory
	//	default template
	//	@name	"template" "other" ...
	SerTemplateDir(name string)
	TemplateDir() string

	//	set template suffix, default ".tpl"
	SetTemplateSuffix(suffix string)
	TemplateSuffix() string

	//	set port
	//	@port	8080|6060...
	SetPort(port int)
	Port() int

	//	set file size upload
	SetFileUploadSize(maxSize int64)
	FileUploadSize() int64

	//	set html encoding type, default utf-8
	SetCharset(encode string)
	Charset() string

	// is ResponseWriter writer compress gizp...
	// According Accept-Encoding select compress type
	// default true
	SetRespWriteCompress(b bool)
	IsRespWriteCompress() bool

	//	set leafveingo http request supported suffixes. e.g.: (.go),(.html) ...
	//	default nil, what form can access, the first is the default suffix
	SetHTTPSuffixs(suffixs ...string)
	HTTPSuffixs() []string

	//	supported resource file suffixs, default ".js", ".css", ".png", ".jpg", ".gif", ".ico", ".html"
	SetStaticFileSuffixs(suffixs ...string)
	StaticFileSuffixs() []string

	//	set server time out default 0. seconds = 10 = 10s
	SetServerTimeout(seconds int64)
	ServerTimeout() int64

	//	http addr default 127.0.0.1
	SetAddr(addr string)
	Addr() string

	//	custom app name. default LeafveingoWeb
	SetAppName(name string)
	AppName() string

	//	custom app version
	SetAppVersion(version string)
	AppVersion() string

	//	leafveingo config
	Config() *Config

	//	log config path
	LogConfPath() string
	SetLogConfPath(path string)

	//	log channel size default 5000
	LogChannelSize() int
	SetLogChannelSize(size int)
}

//	使用Leafvein，会将已经初始化好的ISFLeafvein进行同一返回，程序运行只初始化一次
//	shared init leafveingo, running the program initializes only once
func SharedLeafvein() ISFLeafvein {
	if nil == _thisLeafvein {
		var privatelv sfLeafvein = sfLeafvein{}
		_thisLeafvein = &privatelv
		privatelv.initPrivate()

	}
	return _thisLeafvein
}

//	default ISFLeafvein interface impl
type sfLeafvein struct {
	/* 可重新更改参数 */

	//	app version
	appVersion string

	// http url suffixs
	suffixs []string

	// supported static file suffixs, default ".js", ".css", ".png", ".jpg", ".gif", ".ico", ".html"
	staticFileSuffixs []string

	// html encode type, default utf-8
	charset string

	// is ResponseWriter writer compress gizp...
	// According Accept-Encoding select compress type
	// default true
	isRespWriteCompress bool

	//	file size upload 32M
	fileUploadSize int64

	//	http session maxlife time, unit second
	//	default 1800 second(30 minute)
	sessionMaxlifeTime int32

	/* 需要重新启动程序才能更改参数 */

	// application info default LeafveingoWeb
	appName string

	// server time out, default 0
	serverTimeout int64

	// default 8080
	port int
	// http addr default 127.0.0.1
	addr string

	//	web directory, can read and write to the directory
	//	primary storage html, css, js, image, zip the file
	webRootDir string
	//	template dir, storage template file
	templateDir string

	//	template suffix, default ".tpl"
	templateSuffix string

	isUseSession bool // is use session
	isGCSession  bool // is auto GC session

	//	log config path
	logConfPath string
	//	log channel size default 5000
	logChannelSize int

	/* 对象参数 */

	//	application
	application SFHelper.Map

	//	template
	template LVTemplate.ITemplate

	//	http session manager
	sessionManager *LVSession.HttpSessionManager

	// all controller storage map and router keys
	controllers    map[string]reflect.Value
	controllerKeys []string

	// AdeRouterController interface storage
	controllerArcImpls map[string]AdeRouterController

	//	current leafveingo the file operation directory
	//	operatingDir 是根据 appName 建立的目录路径
	operatingDir string

	//	leafveingo config
	config *Config

	/*	private use */

	//	http url prefix
	prefix string
	//	current http listener
	listener net.Listener
	//	developer mode
	isDevel bool
	//	is start
	isStart bool
}

//	sfLeafvein private init
func (lv *sfLeafvein) initPrivate() {
	lv.controllers = make(map[string]reflect.Value)
	lv.controllerArcImpls = make(map[string]AdeRouterController)
	lv.template = LVTemplate.SharedTemplate()
	lv.application = SFHelper.NewMap()
	lv.isStart = false
	lv.config = new(Config)

	loadConfigByJson([]byte(_defaultConfigJson))

	lv.operatingDir = filepath.Join(SFFileManager.GetExceDir(), lv.appName)

	lv.template.SetBaseDir(lv.TemplateDir())
	lv.template.SetFunc(TEMPLATE_FUNC_KEY_VERSION, lv.Version)
	lv.template.SetFunc(TEMPLATE_FUNC_KEY_APP_NAME, lv.AppName)
	lv.template.SetFunc(TEMPLATE_FUNC_KEY_APP_VERSION, lv.AppVersion)

}

//	主http响应函数
func (lv *sfLeafvein) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	//	TODO 考虑是否加上读取锁，等测试性能后加上再测试看看。

	var context *HttpContext = nil

	defer func() {
		if err := recover(); err != nil {
			var stack string = fmt.Sprintln(err)
			for i := 2; ; i++ {
				pc, file, line, ok := runtime.Caller(i)
				if !ok {
					break
				}

				stack = stack + fmt.Sprintf("%s:%d (0x%x)\n", file, line, pc)
			}

			stack = stack + fmt.Sprintf("\n-----------------------------\nleafveiongo version:%v \ngolang version: %v", lv.Version(), runtime.Version())

			//	暂时字符输出，到时候更改模板
			const ErrorCode500String = "error code 500"
			if nil != context {
				rw.Header().Set("Content-Type", "text/plain; charset="+lv.charset)
				if lv.isDevel {
					context.RespBodyWrite([]byte(stack), HTTP_STATUS_CODE_500)
				} else {
					context.RespBodyWrite([]byte(ErrorCode500String), HTTP_STATUS_CODE_500)
				}
			} else {
				if lv.isDevel {
					http.Error(rw, stack, HTTP_STATUS_CODE_500)
				} else {
					http.Error(rw, ErrorCode500String, HTTP_STATUS_CODE_500)
				}
			}

			lvLog.Error("%v", stack)
		}

		if nil != context {
			context.closeWriter()
		}
	}()

	reqPath := req.URL.Path

	//	前缀url清除，主要使用于集成到别的应用内，定义前缀的去除，以便能正确的访问到自己的应用中
	//	/expand/index = /index
	//	/expand       = /
	if 0 < len(lv.prefix) {
		reqPath = reqPath[len(lv.prefix):]
		if 0 != len(reqPath) && '/' != reqPath[0] {
			reqPath = "/" + reqPath
		}
	}

	if 0 == len(reqPath) {
		reqPath = "/"
	}

	//	静态文件解析
	if reqPath[len(reqPath)-1] != '/' {
		for _, staticSuffixs := range lv.staticFileSuffixs {
			if strings.HasSuffix(reqPath, staticSuffixs) {
				filePath := path.Clean(lv.WebRootDir()) + reqPath

				if isExists, isDir, _ := SFFileManager.Exists(filePath); isExists && !isDir {

					//	处理http.ServeFile函数遇到/index.html被重定向到./的问题
					if strings.HasSuffix(reqPath, INDEX_PAGE) {
						// 防止serveFile做判断，具体可以查看http.ServeFile源码
						req.URL.Path = "/"
					}

					http.ServeFile(rw, req, filePath)

				} else {
					//	404
					http.NotFound(rw, req)
				}
				return
			}
		}
	}

	context = newContext(rw, req, lv.isRespWriteCompress)

	routerKey, methodName, ctrlPath, stuctCode := lv.routerParse(reqPath, context)

	if HTTP_STATUS_CODE_200 == stuctCode {
		var e error = nil
		stuctCode, e = lv.cellController(routerKey, methodName, ctrlPath, context)
		if nil != e {
			lvLog.Error("%v", e)
		}
	}

	switch stuctCode {
	case HTTP_STATUS_CODE_200, HTTP_STATUS_CODE_301, HTTP_STATUS_CODE_307:
		//	不做处理的代码

	//	暂时不处理这些错误代码 TODO
	// case HTTP_STATUS_CODE_503:
	// case HTTP_STATUS_CODE_500:
	// case HTTP_STATUS_CODE_400:
	case HTTP_STATUS_CODE_403:
		const ErrorCode404String = "403 Forbidden The server understood the request"
		rw.Header().Set("Content-Type", "text/plain; charset="+lv.charset)
		context.RespBodyWrite([]byte(ErrorCode404String), HTTP_STATUS_CODE_403)
	// case HTTP_STATUS_CODE_404:
	// 	http.NotFound(rw, req)
	default:
		//	默认跳转404
		const ErrorCode404String = "404 page not found"
		rw.Header().Set("Content-Type", "text/plain; charset="+lv.charset)
		context.RespBodyWrite([]byte(ErrorCode404String), HTTP_STATUS_CODE_404)

	}

}

func (lv *sfLeafvein) start(startName string) {
	SFLog.StartLogManager(lv.logChannelSize)
	SFLog.LoadConfig(lv.logConfPath)

	logInfo := fmt.Sprintf("SFLeafvein %v...\n", startName)

	if 0 == len(lv.controllers) {
		lvLog.Fatal("Leafveingo %v fatal: Controller == nil \n", startName)
		return
	}

	//	目录检测
	if isExites, isDir, _ := SFFileManager.Exists(lv.operatingDir); !isExites || !isDir {
		lvLog.Warn("not locate the operating directory, will not be able to manipulate files,\n %v \n operation directory is created under the app name directory \n", lv.operatingDir)
	}

	if isExites, isDir, _ := SFFileManager.Exists(lv.WebRootDir()); !isExites || !isDir {
		lvLog.Warn("not locate the %v directory, will not be able to read a static file resource and upload file. \n  need to create directory: %v \n", lv.webRootDir, lv.WebRootDir())
	}

	logInfo += "controller:\n"
	//	打印add的控制器
	for key, value := range lv.controllers {
		isArc := ""
		if _, ok := lv.controllerArcImpls[key]; ok {
			isArc = "(Implemented AdeRouterController)"
		}
		logInfo += fmt.Sprintf("%v  \t router : %#v  %v\n", value.Type(), key, isArc)
	}

	//	打印启动信息
	addr := lv.addr
	if 0 < lv.port {
		addr = fmt.Sprintf("%s:%d", lv.addr, lv.port)
	}
	if addr == "" {
		addr = fmt.Sprintf(":%d", 8080)
	}

	//	由于addr设置为127.0.0.1的时候就只能允许内网进行http://localhost:(port)/进行访问，本机IP访问不了。
	//	为了友好的显示，如果addr设置为空的时候允许IP或localhost进行访问做了特别的显示除了（http://0.0.0.0:8080）
	if strings.Index(addr, ":") == 0 {
		logInfo += fmt.Sprintf("Leafveingo %v to listen on %v. Go to http://0.0.0.0%v \n", startName, lv.port, addr)
	} else {
		logInfo += fmt.Sprintf("Leafveingo %v to listen on %v. Go to http://%v \n", startName, lv.port, addr)
	}
	lvLog.Info(logInfo)

	if lv.isUseSession {
		//	自动开启Session GC操作
		lv.sessionManager = LVSession.SharedSessionManager(lv.isGCSession)
	}

	//	设置 server and listen
	server := &http.Server{
		Addr:         addr,
		Handler:      lv,
		ReadTimeout:  time.Duration(lv.serverTimeout) * time.Second,
		WriteTimeout: time.Duration(lv.serverTimeout) * time.Second,
	}

	var err error
	lv.listener, err = net.Listen("tcp", addr)
	if err != nil {
		lvLog.Fatal("Leafveingo %v Listen: %v \n", startName, err)
		return
	}

	lv.isStart = true
	err = server.Serve(lv.listener)
	if err != nil {
		lvLog.Fatal("Leafveingo %v Serve: %v \n", startName, err)
		lv.isStart = false
	}

}

func (lv *sfLeafvein) Start() {

	if !flag.Parsed() {
		flag.Parse()
	}

	lv.isDevel = _flagDeveloper
	lv.template.SetCache(!_flagDeveloper)

	if _flagDeveloper {
		lv.start("DevelStart")
	} else {
		args := flag.Args()
		if 0 < len(args) {
			fmt.Println("incorrect command arguments. \n [-devel] = developer mode, [(nil)] = produce mode.")
			return
		}
		lv.start("Start")
	}

}

func (lv *sfLeafvein) AddController(routerKey string, controller interface{}) {

	//	验证添加路由path的规则
	//	字符串不等于nil || 查询不到"/" || "/" 不在首位
	if len(routerKey) == 0 || routerKey[0] != '/' {
		panic(NewLeafveingoError("%T AddController routerKey path error : %v  reference ( \"/\" | \"/Admin/\" )", controller, routerKey))
	}

	if nil == controller {
		return
	}

	refValue := reflect.ValueOf(controller)
	key := strings.ToLower(routerKey)
	lv.controllers[key] = refValue
	lv.controllerKeys = append(lv.controllerKeys, key)
	sort.Sort(sort.Reverse(SFStringsUtil.SortLengToShort(lv.controllerKeys)))
	//	TODO 排序需要测试

	if refValue.Type().Implements(_arcType) {
		lv.controllerArcImpls[key] = controller.(AdeRouterController)
	}

}

func (lv *sfLeafvein) GetHandlerFunc(prefix string) (handler http.Handler, err error) {
	if 0 == len(lv.controllers) {
		err = errors.New("Leafveingo Start fatal: Controller == nil")
		return
	}

	f := func(w http.ResponseWriter, r *http.Request) {
		//	匹配前缀
		if strings.HasPrefix(r.URL.Path, lv.prefix) {
			lv.ServeHTTP(w, r)
		} else {
			http.NotFound(w, r)
		}
	}
	lv.prefix = prefix
	handler = http.HandlerFunc(f)
	return
}
func (lv *sfLeafvein) Close() {
	if nil != lv.listener {
		err := lv.listener.Close()
		if nil != err {
			lvLog.Fatal("%v", err)
		}
		lvLog.Info("Leafveingo http://%v:%v closed", lv.addr, lv.port)
		lv.isStart = false
	} else {
		lvLog.Fatal("current http listener nil, can not be closed")
	}
}

func (lv *sfLeafvein) IsDevel() bool {
	return lv.isDevel
}
func (lv *sfLeafvein) IsStart() bool {
	return lv.isStart
}
func (lv *sfLeafvein) Application() SFHelper.Map {
	return lv.application
}
func (lv *sfLeafvein) HttpSessionManager() *LVSession.HttpSessionManager {
	return lv.sessionManager
}
func (lv *sfLeafvein) SetUseSession(b bool) {
	lv.isUseSession = b
}
func (lv *sfLeafvein) SetGCSession(b bool) {
	lv.isGCSession = b
}
func (lv *sfLeafvein) SetSessionMaxlifeTime(second int32) {
	if 60 <= second {
		lv.sessionMaxlifeTime = second
	}
}
func (lv *sfLeafvein) SessionMaxlifeTime() int32 {
	return lv.sessionMaxlifeTime
}

func (lv *sfLeafvein) SetHTTPSuffixs(suffixs ...string) {
	for i, v := range suffixs {
		if 0 != len(v) && v[0] != '.' {
			suffixs[i] = "." + v
		}
	}
	lv.suffixs = suffixs
}
func (lv *sfLeafvein) HTTPSuffixs() []string {
	return lv.suffixs
}

func (lv *sfLeafvein) SetStaticFileSuffixs(suffixs ...string) {
	for i, v := range suffixs {
		if 0 != len(v) && v[0] != '.' {
			suffixs[i] = "." + v
		}
	}
	lv.staticFileSuffixs = suffixs
}
func (lv *sfLeafvein) StaticFileSuffixs() []string {
	return lv.staticFileSuffixs
}

func (lv *sfLeafvein) SetTemplateSuffix(suffix string) {
	if 0 != len(suffix) && suffix[0] != '.' {
		suffix = "." + suffix
	}
	lv.templateSuffix = suffix
}
func (lv *sfLeafvein) TemplateSuffix() string {
	return lv.templateSuffix
}

func (lv *sfLeafvein) Version() string {
	return VERSION
}

func (lv *sfLeafvein) SetAddr(addr string) {
	lv.addr = addr
}
func (lv *sfLeafvein) Addr() string {
	return lv.addr
}

func (lv *sfLeafvein) SetPort(port int) {
	if 0 < port {
		lv.port = port
	}
}
func (lv *sfLeafvein) Port() int {
	return lv.port
}

func (lv *sfLeafvein) SetFileUploadSize(maxSize int64) {
	if 0 < maxSize {
		lv.fileUploadSize = maxSize
	}
}
func (lv *sfLeafvein) FileUploadSize() int64 {
	return lv.fileUploadSize
}

func (lv *sfLeafvein) SetCharset(encode string) {
	lv.charset = encode
}
func (lv *sfLeafvein) Charset() string {
	return lv.charset
}

func (lv *sfLeafvein) SetRespWriteCompress(b bool) {
	lv.isRespWriteCompress = b
}
func (lv *sfLeafvein) IsRespWriteCompress() bool {
	return lv.isRespWriteCompress
}

func (lv *sfLeafvein) SetServerTimeout(seconds int64) {
	lv.serverTimeout = seconds
}
func (lv *sfLeafvein) ServerTimeout() int64 {
	return lv.serverTimeout
}

func (lv *sfLeafvein) SetAppName(name string) {
	if !_rexW.MatchString(name) {
		panic(NewLeafveingoError("set AppName format(a-zA-Z0-9) error:%v", name))
	}

	lv.appName = name
	lv.operatingDir = filepath.Join(SFFileManager.GetExceDir(), lv.appName)
	//	由于主操作目录改变，模板目录也需要重新设置主目录
	lv.template.SetBaseDir(lv.TemplateDir())
}
func (lv *sfLeafvein) AppName() string {
	return lv.appName
}

func (lv *sfLeafvein) OperatingDir() string {
	return lv.operatingDir
}
func (lv *sfLeafvein) SetWebRootDir(name string) {
	if !_rexW.MatchString(name) {
		panic(NewLeafveingoError("set WebRootDir format(a-zA-Z0-9) error:%v", name))
	}

	lv.webRootDir = name
}
func (lv *sfLeafvein) WebRootDir() string {
	return filepath.Join(lv.operatingDir, lv.webRootDir)
}

func (lv *sfLeafvein) SerTemplateDir(name string) {
	if !_rexW.MatchString(name) {
		panic(NewLeafveingoError("set TemplateDir format(a-zA-Z0-9) error:%v", name))
	}

	lv.templateDir = name
}
func (lv *sfLeafvein) TemplateDir() string {
	return filepath.Join(lv.operatingDir, lv.templateDir)
}

func (lv *sfLeafvein) SetAppVersion(version string) {
	lv.appVersion = version
}
func (lv *sfLeafvein) AppVersion() string {
	return lv.appVersion
}
func (lv *sfLeafvein) Config() *Config {
	return lv.config
}
func (lv *sfLeafvein) LogConfPath() string {
	return lv.logConfPath
}
func (lv *sfLeafvein) SetLogConfPath(path string) {
	lv.logConfPath = path
}
func (lv *sfLeafvein) LogChannelSize() int {
	return lv.logChannelSize
}
func (lv *sfLeafvein) SetLogChannelSize(size int) {
	if 0 < size {
		lv.logChannelSize = size
	}
}
