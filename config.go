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
//  Update on 2014-07-15
//  Email  slowfei#foxmail.com
//  Home   http://www.slowfei.com

//
//	leafveingo config handle
//
package leafveingo

import (
	"encoding/json"
	"errors"
	"github.com/slowfei/gosfcore/utils/filemanager"
	"io/ioutil"
	"path/filepath"
	"time"
)

const (
	//	default web root dir name
	DEFAULT_WEBROOT_DIR_NAME = "webRoot"
	//	default template dir name
	DEFAULT_TEMPLATE_DIR_NAME = "template"

	//	file size default upload  32M
	DEFAULT_FILE_UPLOAD_SIZE int64 = 32 << 20
	//	server time out default 0
	DEFAULT_SERVER_TIMEOUT = 0
	//	default template suffix
	DEFAULT_TEMPLATE_SUFFIX = ".tpl"
	//	default html charset
	DEFAULT_HTML_CHARSET = "utf-8"
	//	default response write compress
	DEFAULT_RESPONSE_WRITE_COMPRESS = true
	//	default session max life time 1800 seconds
	DEFAULT_SESSION_MAX_LIFE_TIME = 1800
	//	default log channel size
	DEFAULT_LOG_CHANNEL_SIZE = 5000

	//	default KeepAlivePeriod
	DEFAULT_KEEP_ALIVE_PERIOD = 3 * time.Minute
)

var (
	_defaultConfigJson = `
	{
		"AppVersion"			:"1.0",
		"FileUploadSize"		:33554432,
		"Charset"				:"utf-8",
		"StaticFileSuffixes"	:[".js", ".css", ".png", ".jpg", ".gif", ".ico", ".html"],
		"ServerTimeout"			:0,
		"SessionMaxlifeTime"	:1800,
		"IPHeaderKey"			:"",
		"IsReqPathIgnoreCase"	:true,
		"MultiProjectHosts"		:[],

		"TemplateSuffix"		:".tpl",
		"IsCompactHTML"			:true,

		"LogConfigPath"			:"config/log.conf",
		"LogGroup" 				:"",

		"TLSCertPath"			:"",
		"TLSKeyPath"			:"",
		"TLSPort"				:0,
		"TLSAloneRun"			:false,
		
		"IsRespWriteCompress"	:true,
		
		"UserData"				:{}
	}`
)

/**
 *	leafveingo config
 *	default see _defaultConfigJson
 */
type Config struct {
	AppVersion          string   // app version.
	FileUploadSize      int64    // file size upload
	Charset             string   // html encode type
	StaticFileSuffixes  []string // supported static file suffixes
	ServerTimeout       int64    // server time out, default 0
	SessionMaxlifeTime  int32    // http session maxlife time, unit second. use session set
	IPHeaderKey         string   // proxy to http headers set ip key, default ""
	IsReqPathIgnoreCase bool     // request url path ignore case
	MultiProjectHosts   []string // setting integrated multi-project hosts,default nil

	TemplateSuffix string // template suffix
	IsCompactHTML  bool   // is Compact HTML, default true

	LogConfigPath string // log config path, relative or absolute path, relative path from execute file root directory
	LogGroup      string // log group name

	TLSCertPath string // tls cert.pem, relative or absolute path, relative path from execute file root directory
	TLSKeyPath  string // tls key.pem
	TLSPort     int    // tls run prot, default server port+1
	TLSAloneRun bool   // are leafveingo alone run tls server, default false

	// is ResponseWriter writer compress gizp...
	// According Accept-Encoding select compress type
	// default true
	IsRespWriteCompress bool

	UserData map[string]string // user custom config other info
}

/**
 *	conifg load
 *
 *	@param jsonData
 *	@return load error info
 */
func configLoadByJson(jsonData []byte, c *Config) error {

	e2 := json.Unmarshal(jsonData, c)
	if nil != e2 {
		return e2
	}

	return nil
}

/**
 *	conifg load
 *
 *	@param configPath
 *	@return error info
 */
func configLoadByFilepath(configPath string, c *Config) error {

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

	return configLoadByJson(jsonData, c)
}
