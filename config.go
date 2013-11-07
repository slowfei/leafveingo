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
package leafveingo

import (
	"fmt"
	"github.com/slowfei/gosfcore/utils/filemanager"
	"path/filepath"
)

//	Leafveingo config
type Config struct {
	appName             string `json:"AppName"`
	appVersion          string `json:"AppVersion"`
	suffixs             string `json:"Suffixs"`
	staticFileSuffixs   string `json:"StaticFileSuffixs"`
	charset             string
	isRespWriteCompress bool
	fileUploadSize      int64
	sessionMaxlifeTime  int32
	serverTimeout       int64
	port                int
	addr                string
	webRootDir          string
	templateDir         string
	templateSuffix      string
	isUseSession        bool
	isGCSession         bool
	logConfig           string
}

//
//	启动程序前的初始化Leafvein，这个是根据配置文件进行加载的
//
//	@configPath 绝对或相对路径，相对路径从执行文件目录开始查找
//	@return Leafvein的初始化对象，根据配置文件进行配置信息的。
func InitLeafvein(configPath string) ISFLeafvein {
	if 0 == len(configPath) {
		fmt.Println("InitLeafvein config path nil:%v", configPath)
		return nil
	}
	if nil != _thisLeafvein {
		fmt.Println("Leafveingo Has been initialized.")
		return _thisLeafvein
	}

	var path string

	if filepath.IsAbs(configPath) {
		path = configPath
	} else {
		path = SFFileManager.GetExceDir() + configPath
	}

}
