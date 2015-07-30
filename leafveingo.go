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
//  Update on 2015-07-31
//  Email  slowfei#foxmail.com
//  Home   http://www.slowfei.com
//	version 0.0.2.000

//
//	Leafveingo web framework
//
package leafveingo

import (
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/slowfei/gosfcore/debug"
	"github.com/slowfei/gosfcore/helper"
	"github.com/slowfei/gosfcore/log"
	"github.com/slowfei/gosfcore/utils/filemanager"
	"github.com/slowfei/gosfcore/utils/strings"
	"github.com/slowfei/leafveingo/session"
	"github.com/slowfei/leafveingo/template"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (

	//	leafveingo version
	VERSION string = "0.0.2.000"

	// template func key
	TEMPLATE_FUNC_KEY_VERSION     = "Leafveingo_version"
	TEMPLATE_FUNC_KEY_APP_VERSION = "Leafveingo_app_version"
	TEMPLATE_FUNC_KEY_APP_NAME    = "Leafveingo_app_name"
	TEMPLATE_FUNC_KEY_IS_DEVEL    = "Leafveingo_devel"

	//	不知道为什么golang使用http.ServeFile判断为此结尾就重定向到/index.html中。
	//	这里定义主要用来解决这个问题的处理
	INDEX_PAGE = "/index.html"

	//
	FAVICON_PATH = "/favicon.ico"

	//	url params host key
	URL_HOST_KEY = "host"

	//	hosts www
	URL_HOST_WWW     = "www."
	URL_HOST_WWW_LEN = 4

	//	uri schemes
	URI_SCHEME_HTTP  = URIScheme(1) // http
	URI_SCHEME_HTTPS = URIScheme(2) // https
)

var (
	//	开发模式命令
	FlagDeveloper bool

	//	init server memory list
	_serverList []*LeafveinServer = nil

	//
	_serverWaitGroup *sync.WaitGroup = nil
)

func init() {
	flag.BoolVar(&FlagDeveloper, "devel", false, "developer mode start.")
}

//#pragma mark leafveingo func	----------------------------------------------------------------------------------------------------

/**
 *	global start all leafvein server
 *
 */
func Start() {
	if 0 == len(_serverList) {
		SFLog.Fatal("no to start the leafvein server.")
		return
	}

	_serverWaitGroup = new(sync.WaitGroup)

	count := len(_serverList)
	for i := 0; i < count; i++ {
		server := _serverList[i]
		server.Start(true)
	}

	_serverWaitGroup.Wait()
}

/**
 *	global close all leafvein server
 *
 */
func Close() {
	if 0 == len(_serverList) {
		SFLog.Fatal("no to close the leafvein server.")
		return
	}
	count := len(_serverList)
	for i := 0; i < count; i++ {
		server := _serverList[i]
		server.Close()
	}
}

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
		panic(ErrLeafveinAppNameRepeat)
	}
	_serverList = append(_serverList, server)
}

//#pragma mark uri scheme type	----------------------------------------------------------------------------------------------------

//
//	uri scheme
//
type URIScheme int

/**
 *	scheme string
 *
 *	@return string tag
 */
func (u URIScheme) String() string {
	result := ""
	comma := false

	if u&URI_SCHEME_HTTP == URI_SCHEME_HTTP {
		result += "http"
		comma = true
	}

	if u&URI_SCHEME_HTTPS == URI_SCHEME_HTTPS {
		if comma {
			result += ",https"
		} else {
			result += "https"
		}
		comma = true
	}

	return result
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
	Addr string

	Port       int    // ip port. default 8080
	SMGCTime   int64  // session manager gc operate time. 0 is not run session, minimum 60 second. default 300
	ConfigPath string // relative or absolute path, relative path from execute file root directory, can set empty. default ""
}

/**
 *	default server option value
 */
func DefaultOption() ServerOption {
	option := ServerOption{}
	option.Addr = "127.0.0.1"
	option.Port = 8080
	option.SMGCTime = LVSession.DEFAULT_SCAN_GC_TIME
	option.ConfigPath = ""
	return option
}

/**
 *	set addr
 */
func (s ServerOption) SetAddr(addr string) ServerOption {
	s.Addr = addr
	return s
}

/**
 *	set port
 */
func (s ServerOption) SetPort(port int) ServerOption {
	s.Port = port
	return s
}

/**
 *	set session manager operate gc
 */
func (s ServerOption) SetSMGCTime(second int64) ServerOption {
	s.SMGCTime = second
	return s
}

/**
 *	set config file path
 */
func (s ServerOption) SetConfigPath(path string) ServerOption {
	s.ConfigPath = path
	return s
}

/**
 *	checked params
 */
func (s *ServerOption) Checked() {
	if 60 > s.SMGCTime {
		s.SMGCTime = 60
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

	appVersion          string            // app version
	fileUploadSize      int64             // file size upload
	charset             string            // html encode type
	staticFileSuffixes  map[string]string // supported static file suffixes
	serverTimeout       int64             // server time out, default 0
	sessionMaxlifeTime  int32             // http session maxlife time, unit second. use session set
	ipHeaderKey         string            // proxy to http headers set ip key, default ""
	isReqPathIgnoreCase bool              // request url path ignore case
	multiProjectHosts   []string          // setting integrated multi-project hosts

	templateSuffix string // template suffix
	isCompactHTML  bool   // is Compact HTML, default true

	logConfigPath string // log config path, relative or absolute path, relative path from execute file root directory
	logGroup      string // log group name

	tlsCertPath string // tls cert.pem, relative or absolute path, relative path from execute file root directory
	tlsKeyPath  string // tls key.pem
	tlsPort     int    // tls run prot, default server port+1
	tlsAloneRun bool   // are leafveingo alone run tls server

	// is ResponseWriter writer compress gizp...
	// According Accept-Encoding select compress type
	// default true
	isRespWriteCompress bool

	userData map[string]string // user custom config other info

	/* #mark router patams ******************/

	// all router storage map and router keys
	routerList []*RouterElement

	/* #mark system patams ******************/

	application    *SFHelper.Map                 //	application
	sessionManager *LVSession.HttpSessionManager //	http session manager
	template       *LVTemplate.Template          //	template
	log            *SFLog.SFLogger               // log
	config         *Config                       // config info

	//	current leafveingo the file operation directory
	//	operatingDir 是根据 appName 建立的目录路径
	operatingDir string

	memWebRootDir  string       // memory storage
	memTemplateDir string       // memory storage
	prefix         string       // http url prefix
	listener       net.Listener // current http listener
	tlsListener    net.Listener // tls listener
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
	option.Checked()

	server := &LeafveinServer{appName: appName, addr: option.Addr, port: option.Port}

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

	lv.template = LVTemplate.NewTemplate()
	lv.template.SetFunc(TEMPLATE_FUNC_KEY_VERSION, Version)
	lv.template.SetFunc(TEMPLATE_FUNC_KEY_APP_NAME, lv.AppName)
	lv.template.SetFunc(TEMPLATE_FUNC_KEY_APP_VERSION, lv.AppVersion)
	lv.template.SetFunc(TEMPLATE_FUNC_KEY_IS_DEVEL, lv.IsDevel)

	lv.config = new(Config)
	lv.configLoadDefault()

	//	config handle
	if 0 != len(option.ConfigPath) {
		err := configLoadByFilepath(option.ConfigPath, lv.config)
		if nil != err {
			lilcErr := *ErrLeafveinInitLoadConfig
			lilcErr.UserInfo = err.Error()
			panic(&lilcErr)
			return
		}
		lv.configReload(lv.config)
	}

	lv.isStart = false

	lv.application = SFHelper.NewMap()
	lv.operatingDir = filepath.Join(SFFileManager.GetExecDir(), lv.appName)

	//	start session manager
	if 60 <= option.SMGCTime {
		lv.SetHttpSessionManager(LVSession.NewSessionManagerAtGCTime(option.SMGCTime, true))
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

	err := configLoadByJson([]byte(_defaultConfigJson), lv.config)
	if nil != err {
		lldcErr := *ErrLeafveinLoadDefaultConfig
		lldcErr.UserInfo = err.Error()
		panic(&lldcErr)
		return
	}
	lv.configReload(lv.config)
}

/**
 *	reload config info
 *
 *	@param cf
 */
func (lv *LeafveinServer) configReload(cf *Config) {

	lv.SetAppVersion(cf.AppVersion)
	lv.SetFileUploadSize(cf.FileUploadSize)
	lv.SetCharset(cf.Charset)
	lv.SetStaticFileSuffixes(cf.StaticFileSuffixes...)
	lv.SetSessionMaxlifeTime(cf.SessionMaxlifeTime)
	lv.SetIPHeaderKey(cf.IPHeaderKey)
	lv.SetReqPathIgnoreCase(cf.IsReqPathIgnoreCase)
	lv.SetTemplateSuffix(cf.TemplateSuffix)
	lv.SetRespWriteCompress(cf.IsRespWriteCompress)
	lv.SetMultiProjectHosts(cf.MultiProjectHosts...)
	lv.userData = cf.UserData

	//	restart
	lv.SetServerTimeout(cf.ServerTimeout)
	lv.SetCompactHTML(cf.IsCompactHTML)
	lv.SetLogConfigPath(cf.LogConfigPath)
	lv.SetLogGroup(cf.LogGroup)
	lv.SetHttpTLS(cf.TLSCertPath, cf.TLSKeyPath, cf.TLSPort, cf.TLSAloneRun)
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
 *
 *	@param logInfo
 *	@param startName
 *	@return success == true
 */
func (lv *LeafveinServer) parseRouter(logInfo *string, startName string) bool {

	//	memory storage
	lv.memTemplateDir = filepath.Join(lv.operatingDir, DEFAULT_TEMPLATE_DIR_NAME)
	lv.memWebRootDir = filepath.Join(lv.operatingDir, DEFAULT_WEBROOT_DIR_NAME)

	//	log manager
	SFLog.LoadConfig(lv.logConfigPath)
	logTag := fmt.Sprintf("Leafvein(%s)", lv.appName)

	if 0 != len(lv.LogGroup()) {
		lv.log = SFLog.NewLoggerByGroup(logTag, lv.LogGroup())
	} else {
		lv.log = SFLog.NewLogger(logTag)
	}

	//	template
	lv.template.SetBaseDir(lv.TemplateDir())
	lv.template.SetCache(!lv.isDevel)

	//	add global touter
	for _, router := range _globalRouterList {
		if router.appName == lv.AppName() {
			lv.AddRouter(router.router)
		}
	}

	//	validate routerList nil
	if 0 == len(lv.routerList) {
		lv.log.Fatal("LeafveinServer %v fatal: routerList == nil \n", startName)
		return false
	}

	//	validate folder
	if isExites, isDir, _ := SFFileManager.Exists(lv.operatingDir); !isExites || !isDir {
		lv.log.Warn("not locate the operating directory, will not be able to manipulate files,\n %v \n operation directory is created under the app name directory \n", lv.operatingDir)
	}

	if isExites, isDir, _ := SFFileManager.Exists(lv.WebRootDir()); !isExites || !isDir {
		lv.log.Warn("not locate the %v directory, will not be able to read a static file resource and upload file. \n  need to create directory: %v \n", lv.WebRootDir(), lv.WebRootDir())
	}

	//	print config info
	configBuf := bytes.NewBufferString("")
	SFDebug.Fdump(configBuf, true, lv.config)
	*logInfo += "config:\n" + configBuf.String()

	//	print log info
	*logInfo += "controller:\n"

	for _, element := range lv.routerList {
		for _, key := range element.routerKeys {
			if router, ok := element.routers[key]; ok {
				*logInfo += fmt.Sprintf("host:[%#v][%v] key:[%#v] %v\n", element.host, router.ControllerOption().Scheme().String(), key, router.Info())
			} else {
				*logInfo += fmt.Sprintf("schemes[%v][%v] host:[%#v] key:[%#v] controller stores error.", element.host, router.ControllerOption().Scheme().String(), key)
			}
		}
	}

	*logInfo += "\n"

	return true
}

/**
 *	start leafvein server
 *
 *	@param startName "DevelStart" or "Start"
 *	@param goroutine
 */
func (lv *LeafveinServer) start(startName string, goroutine bool) {
	if lv.IsStart() {
		return
	}
	var runFunc func() = nil
	var defaultServer *http.Server = nil

	//	start info
	logInfo := fmt.Sprintf("(%v)Leafvein %v...\n", lv.AppName(), startName)

	if !lv.parseRouter(&logInfo, startName) {
		return
	}

	if 0 >= lv.port {
		lv.port = 8080
	}
	addr := fmt.Sprintf("%s:%d", lv.addr, lv.port)

	//	由于addr设置为127.0.0.1的时候就只能允许内网进行http://localhost:(port)/进行访问，本机IP访问不了。
	//	为了友好的显示，如果addr设置为空的时候允许IP或localhost进行访问做了特别的显示除了（http://0.0.0.0:8080）
	logAddr := ""
	logTLSAddr := ""
	if strings.Index(addr, ":") == 0 {
		logAddr = "http://0.0.0.0" + addr
		logTLSAddr = "https://0.0.0.0"
	} else {
		logAddr = "http://" + addr
		logTLSAddr = "https://" + lv.addr
	}

	if 0 == len(lv.tlsCertPath) || 0 == len(lv.tlsKeyPath) || !lv.tlsAloneRun {
		// default server and listen
		defaultServer = &http.Server{
			Addr:         addr,
			Handler:      lv,
			ReadTimeout:  time.Duration(lv.serverTimeout) * time.Second,
			WriteTimeout: time.Duration(lv.serverTimeout) * time.Second,
		}

		netListen, err := net.Listen("tcp", addr)
		if err != nil {
			lv.log.Fatal("(%v) %v Listen: %v \n", lv.appName, startName, err)
			return
		}
		lv.listener = &leafveinListener{netListen.(*net.TCPListener), DEFAULT_KEEP_ALIVE_PERIOD, nil}
	}

	//	handle tls
	if 0 != len(lv.tlsCertPath) && 0 != len(lv.tlsKeyPath) {
		// see http://golang.org/src/pkg/net/http/server.go?#L1823

		certFullpath := lv.tlsCertPath
		if !filepath.IsAbs(certFullpath) {
			certFullpath = filepath.Join(SFFileManager.GetExecDir(), lv.tlsCertPath)
		}
		keyFullpath := lv.tlsKeyPath
		if !filepath.IsAbs(keyFullpath) {
			keyFullpath = filepath.Join(SFFileManager.GetExecDir(), lv.tlsKeyPath)
		}

		certPEMBlock, err := ioutil.ReadFile(certFullpath)
		if nil != err {
			lv.log.Fatal("(%v) %v Serve Read SSL cert file: %v \n", lv.appName, startName, err)
			return
		}
		keyPEMBlock, err := ioutil.ReadFile(keyFullpath)
		if nil != err {
			lv.log.Fatal("(%v) %v Serve Read SSL key file: %v \n", lv.appName, startName, err)
			return
		}

		if 0 == lv.tlsPort {
			lv.tlsPort = lv.port + 1
		}

		tlsServer := &http.Server{
			Addr:         fmt.Sprintf("%s:%d", lv.addr, lv.tlsPort),
			Handler:      lv,
			ReadTimeout:  time.Duration(lv.serverTimeout) * time.Second,
			WriteTimeout: time.Duration(lv.serverTimeout) * time.Second,
		}

		tlsServer.TLSConfig = new(tls.Config)
		tlsServer.TLSConfig.NextProtos = []string{"http/1.1"}
		tlsServer.TLSConfig.Certificates = make([]tls.Certificate, 1)
		tlsServer.TLSConfig.Certificates[0], err = tls.X509KeyPair(certPEMBlock, keyPEMBlock)
		if nil != err {
			lv.log.Fatal("(%v) %v TLS Serve: %v \n", lv.appName, startName, err)
			return
		}

		tlsListen, err := net.Listen("tcp", tlsServer.Addr)
		if err != nil {
			lv.log.Fatal("(%v) %v TLS Listen: %v \n", lv.appName, startName, err)
			return
		}

		//	http start
		if nil != lv.listener && nil != defaultServer {
			logInfo += fmt.Sprintf("(%v) %v to Listen on %v. Go to %v \n", lv.appName, startName, lv.port, logAddr)
			go func() {
				err = defaultServer.Serve(lv.listener)
				if err != nil {
					lv.log.Fatal("(%v) %v Serve: %v \n", lv.appName, startName, err)
				}
			}()
		}

		//	tls start
		lv.tlsListener = &leafveinListener{tlsListen.(*net.TCPListener), DEFAULT_KEEP_ALIVE_PERIOD, tlsServer.TLSConfig}

		logInfo += fmt.Sprintf("(%v) %v to TLS Listen on %v. Go to %v:%v \n", lv.appName, startName, lv.tlsPort, logTLSAddr, lv.tlsPort)
		lv.log.Info(logInfo)

		runFunc = func() {
			err = tlsServer.Serve(lv.tlsListener)
			if err != nil {
				lv.log.Fatal("(%v) %v TLS Serve: %v \n", lv.appName, startName, err)
			}
			lv.isStart = false
			if nil != _serverWaitGroup {
				_serverWaitGroup.Done()
			}
		}

	} else {

		if nil == defaultServer || nil == lv.listener {
			lv.log.Fatal("default server or default listener init nil. port is configured correctly?")
			return
		}

		logInfo += fmt.Sprintf("(%v) %v to Listen on %v. Go to %v \n", lv.appName, startName, lv.port, logAddr)
		lv.log.Info(logInfo)

		runFunc = func() {
			err := defaultServer.Serve(lv.listener)
			if err != nil {
				lv.log.Fatal("(%v) %v Serve: %v \n", lv.appName, startName, err)
			}
			lv.isStart = false
			if nil != _serverWaitGroup {
				_serverWaitGroup.Done()
			}
		}

	}

	//	start server
	lv.isStart = true
	if nil != _serverWaitGroup {
		_serverWaitGroup.Add(1)
	}

	if goroutine {
		go runFunc()
	} else {
		runFunc()
	}
}

/**
 *	defer func ServeHTTP(...)
 *
 *	@param context
 */
func (lv *LeafveinServer) deferServeHTTP(contextPrt **HttpContext, rw http.ResponseWriter) {

	context := *contextPrt

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

		fmt.Fprintf(stackBuf, "\n-----------------------------\nleafveiongo version:%v \ngolang version: %v", Version(), runtime.Version())
		lv.log.Error("panic error.\nerror info: " + errStr + "\nstack:\n" + stackBuf.String())
	}

	if nil != context {
		context.free()
	}
}

/**
 *	handle request URL
 *
 *	@param rw
 *	@param req
 *	@return reqSuffix	request url suffix
 *	@return reqHost		request url host
 *	@return pass 		is true continue other operations
 */
func (lv *LeafveinServer) requestURLHandle(rw http.ResponseWriter, req *http.Request, reqPath string) (reqSuffix, reqHost string, pass bool) {
	pass = true

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

	//	request suffix, host
	reqSuffix = path.Ext(reqPath)

	hostsCount := len(lv.multiProjectHosts)
	if 0 != hostsCount {
		reqHost = SFStringsUtil.ToLower(req.Host)
		reqHostLen := len(reqHost)

		if 0 == reqHostLen {
			pass = false
			errorMsg := ""
			if lv.IsDevel() {
				errorMsg = "Request Host null"
			}
			lv.statusPageWriter(NewHttpStatusValue(Status400, Status400Msg, errorMsg, ""), rw)
			return
		}

		//	remove port
		retScope := reqHostLen - 7
		if 0 > retScope {
			retScope = 0
		}
		index := -1
		for i := reqHostLen - 1; i >= retScope; i-- {
			if ':' == reqHost[i] {
				index = i
				break
			}
		}
		if -1 != index {
			reqHost = reqHost[:index]
		}

		//	replace 127.0.0.1 host
		if lv.IsDevel() {
			if "localhost" == reqHost || "127.0.0.1" == reqHost {
				urlHost := req.URL.Query().Get(URL_HOST_KEY)
				if 0 != len(urlHost) {
					reqHost = urlHost
				}
			}
		}

		//	remove "www."
		if URL_HOST_WWW_LEN < len(reqHost) && URL_HOST_WWW == reqHost[URL_HOST_WWW_LEN:] {
			reqHost = reqHost[:URL_HOST_WWW_LEN]
		}

		//	checked multi-project host
		pass = false
		for i := 0; i < hostsCount; i++ {
			proHost := lv.multiProjectHosts[i]
			if proHost == reqHost {
				pass = true
				break
			}
		}

		if !pass {
			errorMsg := ""
			if lv.IsDevel() {
				errorMsg = "Host(" + req.Host + ") invalid"
			}
			lv.statusPageWriter(NewHttpStatusValue(Status404, Status404Msg, errorMsg, ""), rw)
			return
		}
	}

	return
}

/**
 *	static file handle
 *
 *	@param rw
 *	@param req
 *	@param reqPath	  request url path
 *	@param reqSuffix  request url suffix
 *	@return pass 	  is true continue other operations
 */
func (lv *LeafveinServer) staticFileHandle(rw http.ResponseWriter, req *http.Request, reqPath, reqSuffix, reqHost string) (pass bool) {
	pass = true

	if _, ok := lv.staticFileSuffixes[reqSuffix]; ok {

		var filePath string

		if 0 != len(reqHost) {
			filePath = lv.WebRootDir() + "/" + reqHost + reqPath
		} else {
			filePath = lv.WebRootDir() + reqPath
		}

		// TODO 考虑使用Join链接路径，然后join后的路径控制在WebRoot目录内的范围，不可越过WebRoot路径。

		if isExists, isDir, _ := SFFileManager.Exists(filePath); isExists && !isDir {
			//	处理http.ServeFile函数遇到/index.html被重定向到./的问题
			if strings.HasSuffix(reqPath, INDEX_PAGE) {
				// 防止serveFile做判断，具体可以查看http.ServeFile源码
				req.URL.Path = "/"
			}
			http.ServeFile(rw, req, filePath)

			pass = false
			return
		} else {
			// favicon.ico 找不到直接跳过
			if FAVICON_PATH == reqPath {
				pass = false
			}

		}
		//	查找不到静态文件交由路由器处理
		// else {
		// 	// 404
		// 	// http.NotFound(rw, req)
		// 	lv.statusPageWriter(NewHttpStatusValue(Status404, Status404Msg, "", ""), rw)
		// }
	}

	return
}

//# mark LeafveinServer public method	--------------------------------------------------------------------------------------------

/**
 *	start leafvein server
 *
 *	@param goroutine true is go func(){ // run serve }(), 关键保留一些操作在主线程来运行, 不知道的可以false
 */
func (lv *LeafveinServer) Start(goroutine bool) {
	if !flag.Parsed() {
		flag.Parse()
	}

	lv.isDevel = FlagDeveloper

	if FlagDeveloper {
		lv.start("DevelStart", goroutine)
	} else {
		args := flag.Args()
		if 0 < len(args) {
			fmt.Println("incorrect command arguments. \n [-devel] = developer mode, [(nil)] = produce mode.")
			return
		}
		lv.start("Start", goroutine)
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
		err = NewLeafveinError("GetHandlerFunc(...) prefix error : %#v  reference ( \"/expand\" )", prefix)
		return
	}

	logInfo := fmt.Sprintf("GetHandlerFunc() parse router...\n")

	if !lv.parseRouter(&logInfo, "HandlerFunc") {
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
 *	testing use
 *
 *	@param parse router result
 */
func (lv *LeafveinServer) TestStart() bool {
	logInfo := ""
	return lv.parseRouter(&logInfo, "TestSet")
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
 * 	@param router
 *
 */
func (lv *LeafveinServer) AddRouter(router IRouter) {

	if lv.isStart {
		lv.log.Warn("AddRouter(...) Leafvein server has been started can not be set.")
		return
	}

	routerKey := router.RouterKey()

	//	验证添加路由path的规则
	//	字符串不等于nil || 查询不到"/" || "/" 不在首位
	if len(routerKey) == 0 || routerKey[0] != '/' {
		panic(NewLeafveinError("AddRouter routerKey error : %#v(%s)  reference ( \"/\" | \"/Admin/\" )", routerKey, router.Info()))
	}

	if nil == router {
		return
	}

	if lv.IsReqPathIgnoreCase() {
		routerKey = SFStringsUtil.ToLower(routerKey)
	}

	controllerHost := router.ControllerOption().Host()
	var element *RouterElement = nil

	for _, elem := range lv.routerList {
		if elem.host == controllerHost {
			element = elem
			break
		}
	}

	if nil == element {
		element = new(RouterElement)
		element.host = controllerHost
		element.routers = make(map[string]IRouter)
		lv.routerList = append(lv.routerList, element)
	}

	if v, ok := element.routers[routerKey]; ok {
		lv.log.Warn("[%#v][%#v]router key already exists(IRouter:%#v), (IRouter:%#v)can not add.", controllerHost, routerKey, v.Info(), router.Info())
		return
	}

	element.routers[routerKey] = router
	element.routerKeys = append(element.routerKeys, routerKey)

	//	由长到短进行排序
	sort.Sort(sort.Reverse(SFStringsUtil.SortLengToShort(element.routerKeys)))
}

/**
 *	add cache template
 *
 *	@param tplName	template unique name
 *	@param src		template content
 *	@return error info
 */
func (lv *LeafveinServer) AddCacheTemplate(tplName, src string) error {
	return lv.template.AddCacheTemplate(tplName, src)
}

/**
 *	add template function
 *
 *	@param `key`		 the template use func key, please ignore repeated system
 *	@param `methodFunc`
 */
func (lv *LeafveinServer) AddTemplateFunc(key string, methodFunc interface{}) {
	lv.template.SetFunc(key, methodFunc)
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
func (lv *LeafveinServer) Application() *SFHelper.Map {
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
		lv.sessionManager.IPHeaderKey = lv.ipHeaderKey
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

	//	由于tlsListener 和 listener 启动的逻辑不一样，所以分别关闭

	if nil != lv.tlsListener && nil != lv.listener {

		err := lv.listener.Close()
		if nil != err {
			lv.log.Fatal("(http) %v", err)
		}
		lv.log.Warn("Leafveingo http://%v:%v closed", lv.addr, lv.port)

		err = lv.tlsListener.Close()
		if nil != err {
			lv.log.Fatal("(tls) %v", err)
		}

		lv.log.Warn("Leafveingo https://%v:%v closed", lv.addr, lv.tlsPort)
		lv.isStart = false
		lv.free()

	} else if nil != lv.tlsListener {

		err := lv.tlsListener.Close()
		if nil != err {
			lv.log.Fatal("(tls) %v", err)
		}

		lv.log.Warn("Leafveingo https://%v:%v closed", lv.addr, lv.tlsPort)
		lv.isStart = false
		lv.free()

	} else if nil != lv.listener {

		err := lv.listener.Close()
		if nil != err {
			lv.log.Fatal("%v", err)
		}
		lv.log.Warn("Leafveingo http://%v:%v closed", lv.addr, lv.port)
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
 *	use reverse proxy set header ip key
 *
 *	@param key
 */
func (lv *LeafveinServer) SetIPHeaderKey(key string) {
	lv.ipHeaderKey = key

	if nil != lv.sessionManager {
		lv.sessionManager.IPHeaderKey = key
	}
}

/**
 *	set request path ignore case
 *
 *	@param ignoreCase default true
 */
func (lv *LeafveinServer) SetReqPathIgnoreCase(ignoreCase bool) {
	if lv.isStart {
		lv.log.Warn("SetReqPathIgnoreCase(...) Leafvein server has been started can not be set.")
		return
	}
	lv.isReqPathIgnoreCase = ignoreCase
}

/**
 *	get is request path ignore case
 *
 *	@return
 */
func (lv *LeafveinServer) IsReqPathIgnoreCase() bool {
	return lv.isReqPathIgnoreCase
}

/**
 *	set multi-project hosts access
 *
 *	@param hosts "slowfei.com"(www.slwofei.com) || "svn.slowfei.com"
 */
func (lv *LeafveinServer) SetMultiProjectHosts(hosts ...string) {
	temp := make([]string, len(hosts))

	for i, host := range hosts {
		if 0 == len(host) {
			panic(ErrLeafveinSetHostNil)
		}

		if URL_HOST_WWW_LEN < len(host) && URL_HOST_WWW == host[URL_HOST_WWW_LEN:] {
			host = host[:URL_HOST_WWW_LEN]
		}
		temp[i] = host
	}

	lv.multiProjectHosts = temp
}

/**
 *	get project hosts
 *
 *	@return
 */
func (lv *LeafveinServer) MultiProjectHosts() []string {
	return lv.multiProjectHosts
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

	for _, v := range suffixes {
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
 *	@param path relative or absolute path, relative path from execute file root directory
 */
func (lv *LeafveinServer) SetLogConfigPath(path string) {
	lv.logConfigPath = path
}

/**
 *	get log config path
 */
func (lv *LeafveinServer) LogConfPath() string {
	return lv.logConfigPath
}

/**
 *	set log group name
 *
 *	Note: need to restart Leafvein Server
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

/**
 *	get log object
 *
 *	@return
 */
func (lv *LeafveinServer) Log() *SFLog.SFLogger {
	return lv.log
}

/**
 *	set tls run params cert.pem and key.pem
 *	Note: need to restart Leafvein Server
 *
 *	@param cretpath  relative or absolute path, relative path from execute file root directory
 *	@param keypath
 *	@param port 	 tls run port
 *	@param aloneRun	 true is alone run tls server
 */
func (lv *LeafveinServer) SetHttpTLS(cretpath, keypath string, port int, aloneRun bool) {
	lv.tlsKeyPath = keypath
	lv.tlsCertPath = cretpath
	lv.tlsPort = port
	lv.tlsAloneRun = aloneRun
}

/**
 *	get tls cert.pem file path
 *
 *	@return
 */
func (lv *LeafveinServer) TLSCertPath() string {
	return lv.tlsCertPath
}

/**
 *	get tls key.pem file path
 *
 *	@return
 */
func (lv *LeafveinServer) TLSKeyPath() string {
	return lv.tlsKeyPath
}

/**
 *	get tls run port
 *
 *	@return
 */
func (lv *LeafveinServer) TLSPort() int {
	return lv.tlsPort
}

//# mark LeafveinHttp override method -------------------------------------------------------------------------------------------

/**
 *	ServeHTTP
 *
 *	@param rw
 *	@param req
 */
func (lv *LeafveinServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	var context *HttpContext = nil
	reqPath := req.URL.Path

	defer lv.deferServeHTTP(&context, rw)

	//	url handle
	reqSuffix, reqHost, urlHandlePass := lv.requestURLHandle(rw, req, reqPath)
	if !urlHandlePass {
		return
	}

	//	static file handle
	if 0 != len(reqSuffix) {
		staticFilePass := lv.staticFileHandle(rw, req, reqPath, reqSuffix, reqHost)
		if !staticFilePass {
			return
		}
	}

	//	create context
	context = newContext(lv, rw, req, lv.isRespWriteCompress)
	context.reqHost = reqHost

	//	router parse
	router, option, statusCode := routerParse(context, reqPath[:len(reqPath)-len(reqSuffix)], reqSuffix)

	//	log info buffer
	logBuf := bytes.NewBuffer([]byte{})
	timeNow := time.Now()

	if nil != option {
		fmt.Fprintf(logBuf, "[%d.%d]request info: (%s)[%s,%s][%s,%s,%d]%#v %#v \n", timeNow.Second(), timeNow.Nanosecond(), lv.AppName(), context.RequestScheme().String(), reqHost, option.RequestMethod, option.RouterKey, statusCode, reqPath, option.RouterPath)
	} else {
		fmt.Fprintf(logBuf, "[%d.%d]request info: (%s)[%s,%s][%s,nil,%d]%#v \n", timeNow.Second(), timeNow.Nanosecond(), lv.AppName(), context.RequestScheme().String(), reqHost, req.Method, statusCode, reqPath)
	}

	errstr := ""
	var err error = nil

	if Status200 == statusCode && nil != router {

		// call controller func and return value handle
		statusCode, err = controllerCallHandle(context, router, option, false, "", logBuf)

		if nil != err {
			if Status200 == statusCode {
				statusCode = Status500
			}
			errstr = err.Error()
		} else {
			switch statusCode {
			case Status200, Status301, Status307, StatusNil:
			default:
				errstr = StatusMsg(statusCode)
			}
		}

		//	保留各个状态的特别处理
		switch statusCode {
		case Status200, Status301, Status307, StatusNil:
			//	不作处理的状态
		default:
			if lv.IsDevel() {
				context.StatusPageWrite(statusCode, StatusMsg(statusCode), errstr, "")
			} else {
				context.StatusPageWrite(statusCode, StatusMsg(statusCode), "", "")
			}
		}

	} else if statusCode != StatusNil {
		if lv.IsDevel() {
			context.StatusPageWrite(statusCode, StatusMsg(statusCode), errstr, "")
		} else {
			context.StatusPageWrite(statusCode, StatusMsg(statusCode), "", "")
		}
	}

	timeNow = time.Now()
	fmt.Fprintf(logBuf, "[%d.%d]status code: (%s)%s \n", timeNow.Second(), timeNow.Nanosecond(), StatusCodeToString(statusCode), errstr)
	lv.log.Info(logBuf.String())
}
