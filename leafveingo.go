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
//  Create on 2013-08-16
//  Update on 2014-06-10
//  Email  slowfei#foxmail.com
//  Home   http://www.slowfei.com
//	version 0.0.2.000

//	web框架leafveingo
package leafveingo

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"github.com/slowfei/gosfcore/helper"
	"github.com/slowfei/gosfcore/log"
	"github.com/slowfei/gosfcore/utils/filemanager"
	"github.com/slowfei/gosfcore/utils/strings"
	"github.com/slowfei/leafveingo/session"
	"github.com/slowfei/leafveingo/template"
	"io"
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
	VERSION string = "0.0.2.000"

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
	TEMPLATE_FUNC_KEY_IS_DEVEL    = "Leafveingo_devel"

	//	不知道为什么golang使用http.ServeFile判断为此结尾就重定向到/index.html中。
	//	这里定义主要用来解决这个问题的处理
	INDEX_PAGE = "/index.html"
)

var (
	//	开发模式命令
	_flagDeveloper bool

	//	TODO Remove
	// _rexW = regexp.MustCompile("\\w+")

	//	init server memory list
	_serverList []*LeafveinServer = nil

	//	temp storage router

)

func init() {
	flag.BoolVar(&_flagDeveloper, "devel", false, "developer mode start.")
}

//#pragma mark leafveingo func	----------------------------------------------------------------------------------------------------

func SetLogManager(channelSize int) {
	// SFLog.StartLogManager(lv.logChannelSize)
	SFLog.StartLogManager(channelSize)
}

/**
 *	by app name get leafvein server
 *
 *	@param appName
 *	@return
 */
func GetServer(appName string) *LeafveinServer {
	count := len(_serverList)
	for i := 0; i < count; i++ {
		server := _serverList[i]
		if appName == server.AppName() {
			return server
		}
	}
	return nil
}

/**
 *	leafveion version info
 *
 *	@return
 */
func Version() string {
	return VERSION
}

/**
 *	private add server
 *
 *	@param server
 */
func addServerList(server *LeafveinServer) {
	if nil != GetServer(server.AppName()) {
		panic(ErrLeafveingoAppNameRepeat)
	}
	_serverList = append(_serverList, server)
}

//#pragma mark ServerOption struct	----------------------------------------------------------------------------------------------------

/**
 *	Leafvein Server option
 *
 *	use DefaultOption() see default value
 */
type ServerOption struct {
	//  ip access 	"127.0.0.1" localhost access
	// 	all address access
	//  "192.168.?.?" 	specify access
	//	default "127.0.0.1"
	addr string

	port       int    // ip port. default 8080
	smgcTime   int64  // session manager gc operate time. 0 is not run session, minimum 60 second. default 300
	configPath string // config file path, can set empty. default ""
}

/**
 *	default server option value
 */
func DefaultOption() ServerOption {
	option := ServerOption{}
	option.addr = "127.0.0.1"
	option.port = 8080
	option.smgcTime = LVSession.DEFAULT_SCAN_GC_TIME
	option.configPath = ""
	return option
}

/**
 *	set addr
 */
func (s *ServerOption) SetAddr(addr string) *ServerOption {
	s.addr = addr
	return s
}

/**
 *	set port
 */
func (s *ServerOption) SetPort(port int) *ServerOption {
	s.port = port
	return s
}

/**
 *	set session manager operate gc
 */
func (s *ServerOption) SetSMGCTime(second int64) *ServerOption {
	s.smgcTime = second
	return s
}

/**
 *	set config file path
 */
func (s *ServerOption) SetConfigPath(path string) *ServerOption {
	s.configPath = path
	return s
}

/**
 *	checked params
 */
func (s *ServerOption) checked() {
	if 60 > s.smgcTime {
		s.smgcTime = 60
	}
}

//# mark LeafveinServer struct	----------------------------------------------------------------------------------------------------

//
//	Leafvein Http Server
//
type LeafveinServer struct {
	/* #mark  require params ******************/

	appName string // application info default LeafveingoWeb
	addr    string // "127.0.0.1" | "" | "192.168.?.?"
	port    int    // 8080 | 8090 ...

	/* #mark optional params, see detailed defaults value to config.go _defaultConfigJson ******************/

	appVersion         string            // app version
	fileUploadSize     int64             // file size upload
	charset            string            // html encode type
	staticFileSuffixes map[string]string // supported static file suffixes
	serverTimeout      int64             // server time out, default 0
	sessionMaxlifeTime int32             // http session maxlife time, unit second. use session set

	templateSuffix string // template suffix
	isCompactHTML  bool   // is Compact HTML, 默认true

	logConfigPath string // log config path
	logGroup      string // log group name

	// is ResponseWriter writer compress gizp...
	// According Accept-Encoding select compress type
	// default true
	isRespWriteCompress bool

	userData map[string]string // user custom config other info

	/* #mark router patams ******************/

	// all router storage map and router keys
	routers    map[string]IRouter
	routerKeys []string

	/* #mark system patams ******************/

	application    SFHelper.Map                  //	application
	sessionManager *LVSession.HttpSessionManager //	http session manager
	template       LVTemplate.ITemplate          //	template
	log            *SFLog.SFLogger

	//	current leafveingo the file operation directory
	//	operatingDir 是根据 appName 建立的目录路径
	operatingDir string

	memWebRootDir  string       // memory storage
	memTemplateDir string       // memory storage
	prefix         string       // http url prefix
	listener       net.Listener // current http listener
	isDevel        bool         // developer mode
	isStart        bool         // is start

}

//# mark LeafveinServer init	----------------------------------------------------------------------------------------------------

/**
 *	new http server
 *
 *	@param appName		application name not is null
 *	@param option		server other option
 */
func NewLeafveinServer(appName string, option ServerOption) *LeafveinServer {
	if 0 == len(appName) {
		return nil
	}
	option.checked()

	server := &LeafveinServer{appName: appName, addr: option.addr, port: option.port}

	//	set other params
	server.initPrivate(option)

	//	add to server list
	addServerList(server)

	return server
}

/**
 *	init optional params
 */
func (lv *LeafveinServer) initPrivate(option ServerOption) {

	server.configLoadDefault()

	//	config handle
	if 0 != len(option.configPath) {
		config, err := configLoadByFilepath(option.configPath)
		if nil != err {
			panic("error init load config: %v", NewLeafveingoError(err.Error()))
			return
		}
		lv.configReload(config)
	}

	lv.isStart = false

	lv.routers = make(map[string]IRouter)
	lv.application = SFHelper.NewMap()
	lv.operatingDir = filepath.Join(SFFileManager.GetExecDir(), lv.appName)

	lv.template = LVTemplate.NewTemplate()
	lv.template.SetFunc(TEMPLATE_FUNC_KEY_VERSION, Version)
	lv.template.SetFunc(TEMPLATE_FUNC_KEY_APP_NAME, lv.AppName)
	lv.template.SetFunc(TEMPLATE_FUNC_KEY_APP_VERSION, lv.AppVersion)
	lv.template.SetFunc(TEMPLATE_FUNC_KEY_IS_DEVEL, lv.IsDevel)

	//	start session manager
	if 60 <= option.smgcTime {
		lv.SetHttpSessionManager(LVSession.NewSessionManagerAtGCTime(option.smgcTime, true))
	}

}

//# mark LeafveinServer private method -------------------------------------------------------------------------------------------

/**
 *	closed free resources
 */
func (lv *LeafveinServer) free() {
	//	TODO Temporarily invalid, keep extension
}

/**
 *	load default config info
 */
func (lv *LeafveinServer) configLoadDefault() {
	config, err := configLoadByJson(_defaultConfigJson)
	if nil != err {
		panic("error load default config: %v", NewLeafveingoError(err.Error()))
		return
	}
	lv.configReload(config)
}

/**
 *	reload config info
 *
 *	@param cf
 */
func (lv *LeafveinServer) configReload(cf Config) {

	lv.SetAppVersion(cf.AppVersion)
	lv.SetFileUploadSize(cf.FileUploadSize)
	lv.SetCharset(cf.Charset)
	lv.SetStaticFileSuffixes(cf.StaticFileSuffixes)
	lv.SetSessionMaxlifeTime(cf.SessionMaxlifeTime)
	lv.SetTemplateSuffix(cf.TemplateSuffix)
	lv.SetRespWriteCompress(cf.IsRespWriteCompress)
	lv.userData = cf.UserData

	//	restart
	lv.SetServerTimeout(cf.ServerTimeout)
	lv.SetCompactHTML(cf.IsCompactHTML)
	lv.SetLogConfigPath(cf.LogConfigPath)
	lv.SetLogGroup(cf.LogGroup)

}

/**
 *	status page response writer set content-type and header code
 *	不使用压缩的输出的
 *
 *	@param rw
 *	@param value
 */
func (lv *LeafveinServer) statusPageWriter(value HttpStatusValue, rw http.ResponseWriter) error {
	rw.Header().Set("Content-Encoding", "none")
	rw.Header().Set("Content-Type", "text/html; charset="+lv.Charset())
	rw.WriteHeader(int(value.status))
	return lv.statusPageExecute(value, rw)
}

/**
 *	status page template execute
 *
 *	@param wr
 *	@param value
 */
func (lv *LeafveinServer) statusPageExecute(value HttpStatusValue, wr io.Writer) error {
	status := strconv.Itoa(int(value.status))

	//	根据状态代码先查找模版，直接查找模版的根目录
	tplName := status + lv.TemplateSuffix()

	tmpl, err := lv.template.Parse(tplName)

	if nil != err {
		tmpl, err = lv.template.ParseString(tplName, HttpStartTemplate)
		if nil != err {
			return err
		}
	}

	return tmpl.Execute(wr, value.data)
}

/**
 *	parse router
 *	TODO
 *
 *	@return success == true
 */
func (lv *LeafveinServer) parseRouter(logInfo *string) bool {

	//	memory storage
	lv.memTemplateDir = filepath.Join(lv.operatingDir, DEFAULT_TEMPLATE_DIR_NAME)
	lv.memWebRootDir = filepath.Join(lv.operatingDir, DEFAULT_WEBROOT_DIR_NAME)

	//	log manager
	SFLog.LoadConfig(lv.logConfigPath)
	logTag := fmt.Sprintf("Leafvein(%s)", lv.appName)
	lv.log = SFLog.NewLogger(logTag)

	//	template
	lv.template.SetBaseDir(lv.TemplateDir())
	lv.template.SetCache(!lv.isDevel)

	//	add global touter
	for _, router := range _globalRouterList {
		if router.appName == lv.AppName() {
			lv.AddRouter(router.routerKey, router.router)
		}
	}

	//	validate routers nil
	if 0 == len(lv.routers) {
		lv.log.Fatal("LeafveinServer %v fatal: routers == nil \n", startName)
		return false
	}

	//	validate folder
	if isExites, isDir, _ := SFFileManager.Exists(lv.operatingDir); !isExites || !isDir {
		lv.log.Warn("not locate the operating directory, will not be able to manipulate files,\n %v \n operation directory is created under the app name directory \n", lv.operatingDir)
	}

	if isExites, isDir, _ := SFFileManager.Exists(lv.WebRootDir()); !isExites || !isDir {
		lv.log.Warn("not locate the %v directory, will not be able to read a static file resource and upload file. \n  need to create directory: %v \n", lv.webRootDir, lv.WebRootDir())
	}

	//	print log info
	*logInfo += "controller:\n"
	for key, value := range lv.routers {
		*logInfo += fmt.Sprintf("[%#v] %v\n", key, value.Info())
	}

	return true
}

/**
 *	start leafvein server
 *
 *	@param startName "DevelStart" or "Start" or "HandlerFunc"
 */
func (lv *LeafveinServer) start(startName string) {

	//	start info
	logInfo := fmt.Sprintf("Leafvein %v...\n", startName)

	if !lv.parseRouter(&logInfo) {
		return
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
	lv.log.Info(logInfo)

	// server and listen
	server := &http.Server{
		Addr:         addr,
		Handler:      lv,
		ReadTimeout:  time.Duration(lv.serverTimeout) * time.Second,
		WriteTimeout: time.Duration(lv.serverTimeout) * time.Second,
	}

	var err error
	lv.listener, err = net.Listen("tcp", addr)
	if err != nil {
		lv.log.Fatal("Leafveingo %v Listen: %v \n", startName, err)
		return
	}

	lv.isStart = true
	err = server.Serve(lv.listener)
	if err != nil {
		lv.log.Fatal("Leafveingo %v Serve: %v \n", startName, err)
		lv.isStart = false
	}
}

/**
 *	defer func ServeHTTP(...)
 *
 *	@param context
 */
func (lv *LeafveinServer) deferServeHTTP(context *HttpContext) {

	if err := recover(); err != nil {
		errStr := fmt.Sprintln(err)

		//	print stack
		stackBuf := bytes.NewBufferString("")
		for i := 2; ; i++ {
			pc, file, line, ok := runtime.Caller(i)
			if !ok {
				break
			}
			fn := runtime.FuncForPC(pc).Name()
			if 0 != len(file) {
				//	/usr/local/go/src/pkg/runtime/proc.c:1223 (0x173d0)
				fmt.Fprintf(stackBuf, "%s(...)\n%s:%d (0x%x)\n", fn, file, line, pc)
			} else {
				// 	runtime.goexit(...)
				// 	L1223: runtime.goexit(...) (0x173d0)
				fmt.Fprintf(stackBuf, "L%d: %s(...) (0x%x)\n", line, fn, pc)
			}
		}

		//	page write
		if nil != context {
			if lv.isDevel {
				context.StatusPageWrite(Status500, Status500Msg, errStr, stackBuf.String())
			} else {
				context.StatusPageWrite(Status500, Status500Msg, "", "")
			}
		} else {
			if lv.isDevel {
				lv.statusPageWriter(NewHttpStatusValue(Status500, Status500Msg, errStr, stackBuf.String()), rw)
			} else {
				lv.statusPageWriter(NewHttpStatusValue(Status500, Status500Msg, "", ""), rw)
			}
		}

		fmt.Fprintf(stackBuf, "\n-----------------------------\nleafveiongo version:%v \ngolang version: %v", lv.Version(), runtime.Version())
		lv.log.Error(stackBuf.String())
	}

	if nil != context {
		context.free()
	}
}

//# mark LeafveinServer public method	--------------------------------------------------------------------------------------------

/**
 *	start leafvein server
 */
func (lv *LeafveinServer) Start() {
	if !flag.Parsed() {
		flag.Parse()
	}

	lv.isDevel = _flagDeveloper

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

/**
 *	获取leafveingo的一个Handler, 方便在其他应用进行扩展leafveingo
 *	get leafveingo handler, Easily be extended to other applications
 *
 *	@prefix	 url prefix
 */
func (lv *LeafveinServer) GetHandlerFunc(prefix string) (handler http.Handler, err error) {

	if 2 > len(prefix) || prefix[0] != '/' {
		err = NewLeafveingoError("GetHandlerFunc(...) prefix error : %#v  reference ( \"/expand\" )", prefix)
		return
	}

	logInfo := fmt.Sprintf("GetHandlerFunc() parse router...\n")

	if !lv.parseRouter(&logInfo) {
		return
	}

	lv.log.Info(logInfo)

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

/**
 *	add router
 *	leafvein interface implement RESTfulRouter ReflectRouter
 *
 *	match rules:
 *		(keys sort of long to short)
 *		keys =[
 *				"/admin/user/"
 *				"/admin/"
 *				"/user/"
 *				"/"
 *			  ]
 *
 *		e.g.:
 *
 *		[urlPath] = [key]; $host = "http://localhost:8080"
 *		$host/admin/user/manager = "/admin/user/"
 *		$host/admin/index		 = "/admin/"
 *		$host/user/index		 = "/user/"
 *		$host/admin              = "/"
 *		$host/index              = "/"
 *
 *
 *	@param routerKey	"/" || "/user/" || "/admin". first char must be '/'
 * 	@param router
 */
func (lv *LeafveinServer) AddRouter(routerKey string, router IRouter) {

	if lv.isStart {
		lv.log.Warn("AddRouter(...) Leafvein server has been started can not be set.")
		return
	}

	//	验证添加路由path的规则
	//	字符串不等于nil || 查询不到"/" || "/" 不在首位
	if len(routerKey) == 0 || routerKey[0] != '/' {
		panic(NewLeafveingoError("%T AddRouter routerKey error : %v  reference ( \"/\" | \"/Admin/\" )", controller, routerKey))
	}

	if nil == router {
		return
	}

	key := strings.ToLower(routerKey)
	if v, ok := lv.routers[key]; ok {
		lv.log.Warn("[%#v]router key already exists(IRouter:%v), (IRouter:%v)can not add.", key, v, router)
		return
	}

	lv.routers[key] = router
	lv.routerKeys = append(lv.routerKeys, key)

	//	由长到短进行排序
	sort.Sort(sort.Reverse(SFStringsUtil.SortLengToShort(lv.routerKeys)))
}

/**
 *	developer mode
 *
 *	@return true is developer mode
 */
func (lv *LeafveinServer) IsDevel() bool {
	return lv.isDevel
}

/**
 *	is starting
 *
 *	@return
 */
func (lv *LeafveinServer) IsStart() bool {
	return lv.isStart
}

/**
 *	web application map
 *
 *	@return
 */
func (lv *LeafveinServer) Application() SFHelper.Map {
	return lv.application
}

/**
 *	custom set session manager
 *
 *	@param sessionManager nil == stop use session
 */
func (lv *LeafveinServer) SetHttpSessionManager(sessionManager *LVSession.HttpSessionManager) {

	if lv.isStart {
		lv.log.Warn("SetHttpSessionManager(...) Leafvein server has been started can not be set.")
		return
	}

	if nil != lv.sessionManager {
		lv.sessionManager.Free()
		lv.sessionManager = nil
	}

	if nil != sessionManager {
		lv.sessionManager = sessionManager
	}
}

/**
 *	http session manager
 */
func (lv *LeafveinServer) HttpSessionManager() *LVSession.HttpSessionManager {
	return lv.sessionManager
}

/**
 *	user custom config other info
 *
 *	@param key
 *	@return
 */
func (lv *LeafveinServer) UserData(key string) string {
	return lv.userData[key]
}

/**
 *	current leafveingo the file operation directory
 *	OperatingDir() path is created based on AppName
 *
 *	@return
 */
func (lv *LeafveinServer) OperatingDir() string {
	return lv.operatingDir
}

/**
 *	template directory, primary storage template file
 *	OperatingDir() created under the directory
 *
 *	../sample(AppName)/template
 */
func (lv *LeafveinServer) TemplateDir() string {
	return lv.memTemplateDir
}

/**
 *	web directory, can read and write to the directory
 *	primary storage html, css, js, image, zip the file
 *	OperatingDir() created under the directory
 *
 *	../sample(AppName)/webRoot
 */
func (lv *LeafveinServer) WebRootDir() string {
	return lv.memWebRootDir
}

/**
 *	close Leafvein server
 *
 *	TODO Temporarily invalid, keep extension
 */
func (lv *LeafveinServer) Close() {
	if nil != lv.listener {
		err := lv.listener.Close()
		if nil != err {
			lv.log.Fatal("%v", err)
		}
		lv.log.Info("Leafveingo http://%v:%v closed", lv.addr, lv.port)
		lv.isStart = false
		lv.free()
	} else {
		lv.log.Fatal("current http listener nil, can not be closed")
	}
}

//# mark LeafveinServer get set method -------------------------------------------------------------------------------------------

/**
 *	get application name
 *
 *	@return
 */
func (lv *LeafveinServer) AppName() string {
	return lv.appName
}

/**
 *	get addr
 *
 *	@return
 */
func (lv *LeafveinServer) Addr() string {
	return lv.addr
}

/**
 *	get port
 *
 *	@return
 */
func (lv *LeafveinServer) Port() int {
	return lv.port
}

/**
 *	set server time out
 *
 *	Note: need to restart Leafvein Server
 *
 *	@param seconds default 0. seconds = 10 = 10s
 */
func (lv *LeafveinServer) SetServerTimeout(seconds int64) {
	lv.serverTimeout = seconds
}

/**
 *	get server time out
 *
 *	@return
 */
func (lv *LeafveinServer) ServerTimeout() int64 {
	return lv.serverTimeout
}

/**
 *	set application version
 *
 *	@param version
 */
func (lv *LeafveinServer) SetAppVersion(version string) {
	lv.appVersion = version
}

/**
 *	get application version
 *
 *	@return
 */
func (lv *LeafveinServer) AppVersion() string {
	return lv.appVersion
}

/**
 *	set http session maxlife time, unit second
 *
 *	@param second default 1800 second(30 minute), must be greater than 60 seconds
 */
func (lv *LeafveinServer) SetSessionMaxlifeTime(second int32) {

	if 60 > second {
		second = 60
	}
	lv.sessionMaxlifeTime = second
}

/**
 *	get http session maxlife time, unit second
 */
func (lv *LeafveinServer) SessionMaxlifeTime() int32 {
	return lv.sessionMaxlifeTime
}

/**
 *	set static resource file suffixes
 *
 *	@param suffixes default ".js", ".css", ".png", ".jpg", ".gif", ".ico", ".html"
 */
func (lv *LeafveinServer) SetStaticFileSuffixes(suffixes ...string) {

	if lv.isStart {
		lv.log.Warn("SetStaticFileSuffixs(...) Leafvein server has been started can not be set.")
		return
	}

	lv.staticFileSuffixes = make(map[string]string)

	for i, v := range suffixes {
		if 0 != len(v) {
			key := v
			if key[0] != '.' {
				key = "." + key
			}
			lv.staticFileSuffixes[key] = key
		}
	}
}

/**
 *	get static resource file suffixs
 *
 *	@return
 */
func (lv *LeafveinServer) StaticFileSuffixes() []string {

	tempSlice := make([]string, len(lv.staticFileSuffixes))

	i := 0
	for k, _ := range lv.staticFileSuffixes {
		tempSlice[i] = k
		i++
	}

	return tempSlice
}

/**
 *	set template suffix,
 *
 *	@param suffix default ".tpl"
 */
func (lv *LeafveinServer) SetTemplateSuffix(suffix string) {
	if 0 != len(suffix) && suffix[0] != '.' {
		suffix = "." + suffix
	}
	lv.templateSuffix = suffix
}

/**
 *	get template suffix
 *
 *	@return
 */
func (lv *LeafveinServer) TemplateSuffix() string {
	return lv.templateSuffix
}

/**
 *	set file size upload
 *
 *	@param maxSize default 32M
 */
func (lv *LeafveinServer) SetFileUploadSize(maxSize int64) {
	if 0 < maxSize {
		lv.fileUploadSize = maxSize
	}
}

/**
 *	get file size upload
 *
 *	@return
 */
func (lv *LeafveinServer) FileUploadSize() int64 {
	return lv.fileUploadSize
}

/**
 *	set html encoding type
 *
 *	@param encode default utf-8
 */
func (lv *LeafveinServer) SetCharset(encode string) {
	lv.charset = encode
}

/**
 *	get html encoding type
 *
 *	@return
 */
func (lv *LeafveinServer) Charset() string {
	return lv.charset
}

/**
 *	is ResponseWriter writer compress gizp...
 *	According Accept-Encoding select compress type
 *
 *  Note: need to restart Leafvein Server
 *
 *	@param b default true
 */
func (lv *LeafveinServer) SetRespWriteCompress(b bool) {
	lv.isRespWriteCompress = b
}

/**
 *	RespWriteCompress
 *
 *	@return
 */
func (lv *LeafveinServer) IsRespWriteCompress() bool {
	return lv.isRespWriteCompress
}

/**
 *	out html is compact remove (\t \n space) sign
 *	only template file out use
 *
 *  Note: need to restart Leafvein Server
 *
 *	@param compact default true
 */
func (lv *LeafveinServer) SetCompactHTML(compact bool) {
	lv.template.SetCompactHTML(compact)
}

/**
 *	get is compact HTML
 */
func (lv *LeafveinServer) IsCompactHTML() bool {
	return lv.template.IsCompactHTML()
}

/**
 *	set log config path
 *
 *  Note: need to restart Leafvein Server
 *
 *	@param path
 */
func (lv *LeafveinServer) SetLogConfigPath(path string) {
	lv.logConfPath = path
}

/**
 *	get log config path
 */
func (lv *LeafveinServer) LogConfPath() string {
	return lv.logConfPath
}

/**
 *	set log group name
 *
 *	Note: need to restart application
 *
 *	@param groupName
 */
func (lv *LeafveinServer) SetLogGroup(groupName string) {
	lv.logGroup = groupName
}

/**
 *	get log group name
 *
 *	@return
 */
func (lv *LeafveinServer) LogGroup() string {
	return lv.logGroup
}

//# mark LeafveinHttp override method -------------------------------------------------------------------------------------------

/**
 *	ServeHTTP
 */
func (lv *LeafveinServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	//	TODO 考虑是否加上读取锁，等测试性能后加上再测试看看。

	var context *HttpContext = nil
	reqPath := req.URL.Path

	defer lv.deferServeHTTP(context)

	//	前缀url清除，call GetHandlerFunc() 才会使用到此操作
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

	//	get url ext suffix
	reqSuffix := path.Ext(reqPath)

	//	static file handle
	if 0 != len(reqSuffix) {
		if _, ok := lv.staticFileSuffixes[reqSuffix]; ok {
			filePath := lv.WebRootDir() + reqPath

			if isExists, isDir, _ := SFFileManager.Exists(filePath); isExists && !isDir {
				//	处理http.ServeFile函数遇到/index.html被重定向到./的问题
				if strings.HasSuffix(reqPath, INDEX_PAGE) {
					// 防止serveFile做判断，具体可以查看http.ServeFile源码
					req.URL.Path = "/"
				}
				http.ServeFile(rw, req, filePath)
			} else {
				// 404
				// http.NotFound(rw, req)
				lv.statusPageWriter(NewHttpStatusValue(Status404, Status404Msg, "", ""), rw)
			}
			return
		}
	}

	lv.log.Info("request url path: %#v", reqPath)

	//	create context
	context = newContext(lv, rw, req, lv.isRespWriteCompress)

	//	router parse
	router, option, statusCode := routerParse(context, reqPath[:len(reqPath)-len(reqSuffix)], reqSuffix)

	if Status200 == statusCode && nil != router {
		errstr := ""
		var err error = nil
		funcName := ""
		tplPath := ""

		//
		funcName, tplPath, statusCode, err = router.ParseController(context, option)

		if Status200 == statusCode {
			//TODO

		} else if nil != err {
			errors = err.Error()
		} else {
			errstr = "Unknown Error."
		}

		switch statusCode {
		case Status200, Status301, Status307:
			//	不作处理的状态
		case Status400:
			context.StatusPageWrite(statusCode, Status400Msg, errstr, "")
		case Status403:
			context.StatusPageWrite(statusCode, Status403Msg, errstr, "")
		case Status500:
			context.StatusPageWrite(statusCode, Status500Msg, errstr, "")
		case Status503:
			context.StatusPageWrite(statusCode, Status503Msg, errstr, "")
		default:
			context.StatusPageWrite(statusCode, Status404Msg, errstr, "")
		}

	} else {
		context.StatusPageWrite(statusCode, StatusMsg(statusCode), "", "")
	}

}
