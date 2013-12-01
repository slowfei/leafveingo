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
//  Create on 2013-11-06
//  Update on 2013-11-06
//  Email  slowfei@foxmail.com
//  Home   http://www.slowfei.com

//	leafveingo文件配置加载与初始化
//
//
package leafveingo

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/slowfei/gosfcore/utils/filemanager"
	"io/ioutil"
	"path/filepath"
)

const (
	//  default port 8080
	DEFAULT_PORT int = 8080
	//	default http addr
	DEFAULT_HTTP_ADDR = "127.0.0.1"
	//	file size default upload  32M
	DEFAULT_FILE_UPLOAD_SIZE int64 = 32 << 20
	//	server time out default 0
	DEFAULT_SERVER_TIMEOUT = 0
	//	default web root dir
	DEFAULT_WEBROOT_DIR = "webRoot"
	//	default template dir
	DEFAULT_TEMPLATE_DIR = "template"
	//	default template suffix
	DEFAULT_TEMPLATE_SUFFIX = ".tpl"
	//	default html charset
	DEFAULT_HTML_CHARSET = "utf-8"
	//	default response write compress
	DEFAULT_RESPONSE_WRITE_COMPRESS = true
	//	default session max life time 1800 seconds
	DEFAULT_SESSION_MAX_LIFE_TIME = 1800
	//	default use session
	DEFAULT_SESSION_USE = true
	//	default gc session
	DEFAULT_SESSION_GC = true
	//	default log channel size
	DEFAULT_LOG_CHANNEL_SIZE = 5000
)

var (
	_defaultConfigJson = `
	{
		"Port"			:8080,
		"Addr"			:"127.0.0.1",
		"ServerTimeout"		:0,
		"AppName"		:"LeafveingoWeb",
		"AppVersion"		:"1.0",
		"Suffixs"		:[""],
		"StaticFileSuffixs"	:[".js", ".css", ".png", ".jpg", ".gif", ".ico", ".html"],
		"Charset"		:"utf-8",
		"IsRespWriteCompress"	:true,
		"FileUploadSize"	:33554432,
		"IsUseSession"		:true,
		"IsGCSession"		:true,
		"SessionMaxlifeTime"	:1800,
		"WebRootDir"		:"webRoot",
		"TemplateDir"		:"template",
		"TemplateSuffix"	:".tpl",
		"LogConfigPath"		:"config/log.conf",
		"LogChannelSize"	:5000,
		"UserData"		:{}
	}`
)

//	Leafveingo config
type Config struct {
	Port                int               // default port 8080
	Addr                string            // default http addr 127.0.0.1
	ServerTimeout       int64             // server time out default 0
	AppName             string            // app name default "LeafveingoWeb"
	AppVersion          string            // app version default 1.0
	Suffixs             []string          // http request url suffixs default ""
	StaticFileSuffixs   []string          // static file suffixs default ".js", ".css", ".png", ".jpg", ".gif", ".ico", ".html"
	Charset             string            // default html charset utf-8
	IsRespWriteCompress bool              // default response write compress true
	FileUploadSize      int64             // file size default upload  32M
	IsUseSession        bool              // default use http session true
	IsGCSession         bool              // default gc http session true
	SessionMaxlifeTime  int32             // default session max life time 1800 seconds
	WebRootDir          string            // default web root dir "webRoot"
	TemplateDir         string            // default template dir "template"
	TemplateSuffix      string            // default template suffix ".tpl"
	LogConfigPath       string            // default log config path relative path"config/app.conf"
	LogChannelSize      int               // default log channel size 5000
	UserData            map[string]string //
}

func (c *Config) Get(key string) string {
	return c.UserData[key]
}

//	load config
//	@configPath
func loadConfig(configPath string) error {
	if nil == _thisLeafvein {
		return ErrLeafveingoNotInit
	}
	if nil == _thisLeafvein.Config() {
		return ErrLeafveingoConfigNotInit
	}

	if 0 != len(configPath) {
		// _rwmutex.Lock()
		// defer _rwmutex.Unlock()

		var path string
		if filepath.IsAbs(configPath) {
			path = configPath
		} else {
			path = filepath.Join(SFFileManager.GetExecDir(), configPath)
		}

		isExists, isDir, _ := SFFileManager.Exists(path)
		if !isExists || isDir {
			return errors.New("failed to load configuration file:" + configPath)
		}

		jsonData, e1 := ioutil.ReadFile(path)
		if nil != e1 {
			return e1
		}

		e2 := json.Unmarshal(jsonData, _thisLeafvein.Config())

		if nil != e2 {
			return e2
		}
	}
	setLeafveingoConfig()
	return nil
}

//	load config
//	@jsonData
func loadConfigByJson(jsonData []byte) error {
	if nil == _thisLeafvein {
		return ErrLeafveingoNotInit
	}

	if nil == _thisLeafvein.Config() {
		return ErrLeafveingoConfigNotInit
	}

	e2 := json.Unmarshal(jsonData, _thisLeafvein.Config())
	if nil != e2 {
		return e2
	}
	setLeafveingoConfig()
	return nil
}

//
//	启动程序前的初始化Leafvein，这个是根据配置文件进行加载的
//
//	@configPath 绝对或相对路径，相对路径从执行文件目录开始查找
//	@return Leafvein的初始化对象，根据配置文件进行配置信息的。
func InitLeafvein(configPath string) ISFLeafvein {
	if nil != _thisLeafvein {
		panic(ErrLeafveingoHasbeenInit)
	}

	var privatelv sfLeafvein = sfLeafvein{}
	_thisLeafvein = &privatelv
	privatelv.initPrivate()

	if 0 != len(configPath) {
		err := loadConfig(configPath)
		if nil != err {
			panic(NewLeafveingoError(err.Error()))
		}
	} else {
		fmt.Println("InitLeafvein Failed to load config file: ", configPath)
		fmt.Println("Use default config info:\n", _defaultConfigJson)
	}

	return _thisLeafvein
}

//	进行配置的设置。
func setLeafveingoConfig() {
	if nil != _thisLeafvein {
		cf := _thisLeafvein.Config()
		if nil != cf {
			if !_thisLeafvein.IsStart() {
				//	启动后不能修改的配置信息

				// appName string
				_thisLeafvein.SetAppName(cf.AppName)

				// server time out
				_thisLeafvein.SetServerTimeout(cf.ServerTimeout)

				// default 8080
				_thisLeafvein.SetPort(cf.Port)

				// http addr default 127.0.0.1
				// addr string
				_thisLeafvein.SetAddr(cf.Addr)

				// web directory, can read and write to the directory
				// primary storage html, css, js, image, zip the file
				// webRootDir string
				_thisLeafvein.SetWebRootDir(cf.WebRootDir)

				// template dir, storage template file
				// templateDir string
				_thisLeafvein.SerTemplateDir(cf.TemplateDir)

				// template suffix, default ".tpl"
				// templateSuffix string
				_thisLeafvein.SetTemplateSuffix(cf.TemplateSuffix)

				// isUseSession bool // is use session
				_thisLeafvein.SetUseSession(cf.IsUseSession)

				// isGCSession  bool // is auto GC session
				_thisLeafvein.SetGCSession(cf.IsGCSession)

				// log config path
				// logConfPath string
				_thisLeafvein.SetLogConfPath(cf.LogConfigPath)

				// log channel size default 5000
				// logChannelSize int
				_thisLeafvein.SetLogChannelSize(cf.LogChannelSize)

			}

			// 	appVersion string
			_thisLeafvein.SetAppVersion(cf.AppVersion)

			// http url suffixs
			_thisLeafvein.SetHTTPSuffixs(cf.Suffixs...)

			// supported static file suffixs
			_thisLeafvein.SetStaticFileSuffixs(cf.StaticFileSuffixs...)

			// html encode type, charset
			_thisLeafvein.SetCharset(cf.Charset)

			// isRespWriteCompress bool
			_thisLeafvein.SetRespWriteCompress(cf.IsRespWriteCompress)

			// file size upload 32M fileUploadSize
			_thisLeafvein.SetFileUploadSize(cf.FileUploadSize)

			// sessionMaxlifeTime int32
			_thisLeafvein.SetSessionMaxlifeTime(cf.SessionMaxlifeTime)

		}
	}
}
