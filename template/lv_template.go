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
//  Create on 2013-8-24
//  Update on 2013-10-23
//  Email  slowfei@foxmail.com
//  Home   http://www.slowfei.com

//	leafveingo web 模板操作模块
package LVTemplate

import (
	"errors"
	"github.com/slowfei/gosfcore/utils/filemanager"
	"github.com/slowfei/gosfcore/utils/time"
	"html/template"
	"io"
	"io/ioutil"
	"path"
	"runtime"
	"sync"
)

const (
	//嵌入模板函数key名
	kLVEmbedTempate = "LVEmbedTempate"
	kLVMapPack      = "LVMapPack"
	kLVTimeFormat   = "LVTimeFormat"
	kLVStringToHtml = "LVStringToHtml"
	kGOVersion      = "GoVersion"
)

var (
	// private
	thisTemplate ITemplate
)

//	模板数据，用于封装使用模板的数据传递
type TemplateValue struct {
	TplPath string      //	模板的相对路径
	Data    interface{} //	模板绑定的数据
}

// 获取模板对象
func SharedTemplate() ITemplate {
	if nil == thisTemplate {
		t := lvtemplate{}
		t.initPrivate()
		thisTemplate = &t
	}
	return thisTemplate
}

//	new template value
//	@tplPth		 相对路径
//	@data		  模板数据
func NewTemplateValue(tplPath string, data interface{}) TemplateValue {
	return TemplateValue{tplPath, data}
}

//	new template value
//	@data		  模板数据
func NewTemplateValueByData(data interface{}) TemplateValue {
	return TemplateValue{Data: data}
}

//	模板使用接口
type ITemplate interface {
	//	设置模板函数操作
	//	请在初始化程序时添加，如果已经缓存的模板不会进行处理
	//	内置默认函数(具体可查看lv_tempate_func.go)：
	//			"LVEmbedTempate" LVTemplate.EmbedTempate	//嵌入模板函数
	//	@key
	//	@methodFunc
	SetFunc(key string, methodFunc interface{})
	DelFunc(key string)
	DelAllFunc()

	//	设置模板查询主目录,默认"".
	//	主目录空的话就会按照编译文件所在目录开始查询模板
	//	@pathDir	完整路径目录
	SetBaseDir(pathDir string)
	BaseDir() string

	//	根据相对路径获取模板的完整路径，如果查询不到模板返回""
	//	@tplPath	模板的相对路径
	TemplatePathAtName(tplPath string) string

	//	写入模板操作
	//	@wr
	//	@value
	Execute(wr io.Writer, value TemplateValue) error

	//	分析模板
	//	返回的模板是经过New()->Funcs()->Parse()处理后的
	//	@tplPath	模板的相对路径
	//	@return
	Parse(tplPath string) (*template.Template, error)

	//	根据模板相对路径获取模板，如果查询不到返回 nil
	//	@tplPath	模板的相对路径
	Get(tplPath string) *template.Template

	//	是否设置模板缓存处理
	SetCache(cache bool)
	IsCache() bool
}

//	leafveingo 内置模板结构，用于私有的实现
type lvtemplate struct {
	funcMap    template.FuncMap   // 模板函数
	goTemplate *template.Template // 唯一的模板管理
	isCache    bool               // 是否加入缓存
	baseDir    string             // 模板路径的主目录
	rwmutex    sync.RWMutex
}

func (l *lvtemplate) initPrivate() {
	l.funcMap = make(template.FuncMap)
	l.funcMap[kLVEmbedTempate] = l.embedTempate
	l.funcMap[kLVStringToHtml] = l.stringToHtml
	l.funcMap[kLVMapPack] = l.mapPack
	l.funcMap[kGOVersion] = runtime.Version
	l.funcMap[kLVTimeFormat] = SFTimeUtil.YMDHMSSFormat
	l.goTemplate = template.New("LVTemplate")
	l.baseDir = ""
}

func (l *lvtemplate) Get(tplPath string) *template.Template {
	return l.goTemplate.Lookup(tplPath)
}

func (l *lvtemplate) Parse(tplPath string) (*template.Template, error) {

	var tmpl *template.Template
	if l.isCache {
		if tmpl = l.goTemplate.Lookup(tplPath); nil != tmpl {
			return tmpl, nil
		}
	}

	if fullPath := l.TemplatePathAtName(tplPath); 0 < len(fullPath) {
		read, err := ioutil.ReadFile(fullPath)
		if nil != err {
			return nil, err
		}
		tplString := string(read)

		if l.isCache {
			tmpl = l.goTemplate.New(tplPath)
		} else {
			tmpl = template.New(tplPath)
		}

		tmpl.Funcs(l.funcMap)
		if _, err := tmpl.Parse(tplString); nil != err {
			return nil, err
		}
		return tmpl, nil
	} else {
		return nil, errors.New("can not find :" + tplPath)
	}

}

func (l *lvtemplate) Execute(wr io.Writer, value TemplateValue) error {
	tmpl, err := l.Parse(value.TplPath)
	if nil == err && nil != tmpl {
		tmpl.Execute(wr, value.Data)
		return nil
	} else {
		return err
	}
}

func (l *lvtemplate) TemplatePathAtName(tplPath string) string {
	var fullPath string

	if 0 < len(l.baseDir) {
		fullPath = path.Join(l.baseDir, tplPath)
	} else {
		fullPath = path.Join(SFFileManager.GetExceDir(), tplPath)
	}

	isDir := false
	if b, _ := SFFileManager.Exists(fullPath, &isDir); b && !isDir {
		return fullPath
	}

	return ""
}

func (l *lvtemplate) SetFunc(key string, methodFunc interface{}) {
	l.rwmutex.Lock()
	defer l.rwmutex.Unlock()
	l.funcMap[key] = methodFunc
}

func (l *lvtemplate) DelFunc(key string) {
	l.rwmutex.Lock()
	defer l.rwmutex.Unlock()

	delete(l.funcMap, key)
}

func (l *lvtemplate) DelAllFunc() {
	l.rwmutex.Lock()
	defer l.rwmutex.Unlock()

	for k, _ := range l.funcMap {
		delete(l.funcMap, k)
	}
}

func (l *lvtemplate) SetBaseDir(path string) {
	l.rwmutex.Lock()
	defer l.rwmutex.Unlock()

	l.baseDir = path
}
func (l *lvtemplate) BaseDir() string {
	l.rwmutex.RLock()
	defer l.rwmutex.RUnlock()
	return l.baseDir
}

func (l *lvtemplate) SetCache(cache bool) {
	l.rwmutex.Lock()
	defer l.rwmutex.Unlock()

	l.isCache = cache
}
func (l *lvtemplate) IsCache() bool {
	l.rwmutex.RLock()
	defer l.rwmutex.RUnlock()

	return l.isCache
}
