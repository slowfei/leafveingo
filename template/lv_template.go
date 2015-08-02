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
//  Update on 2015-08-03
//  Email  slowfei#nnyxing.com
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
	"regexp"
	"runtime"
	"sync"
)

const (
	//嵌入模板函数key名
	kLVEmbedTempate       = "LVEmbedTempate"
	kLVEmbedTempateByName = "LVEmbedTempateByName"
	kLVMapPack            = "LVMapPack"
	kLVTimeFormat         = "LVTimeFormat"
	kLVStringToHtml       = "LVStringToHtml"
	kGOVersion            = "GoVersion"
)

var (

	/*  根据顺序使用正则来去除html的换行和空格
	TODO 在javascript中括号中的空格还未完全去除，此正则操作可能过于繁琐，还没有找到好的解决方案。
	而且也比较耗时间，不过在进行缓存后可以忽略
	*/
	//	1.标签外的空格
	_rexTagExtSpace = regexp.MustCompile(`>[[:space:]]+<`)
	//	2.去除空行与注释
	_rexEmptyLineNode = regexp.MustCompile(`(\n\s+|\n)|(<%--[\s\S]+?-->)`)
	//	3.去除标签内的首个空格
	_rexTagFirstSpace = regexp.MustCompile(`(<\w+)\s{2,}`)
	//	4.去除标签内结尾空格
	_rexTagLastSpace = regexp.MustCompile(`\s+(>|/>)`)
	//	5.去除"="符号的两边空格
	_rexSignSpace = regexp.MustCompile(`\s*(=|!=|==)\s*`)
)

//	模板数据，用于封装使用模板的数据传递
type TemplateValue struct {
	TplName     string      //	template name precedence handle
	TplPath     string      //	template relative path
	ContentType string      //	response header Content-Type
	Data        interface{} //	bind data
}

// 获取模板对象
func NewTemplate() *Template {
	t := new(Template)
	t.initPrivate()
	return t
}

//	new template value
//	@tplPth		 相对路径
//	@data		  模板数据
func NewTemplateValue(tplPath string, data interface{}) TemplateValue {
	return TemplateValue{TplPath: tplPath, Data: data}
}

/**
 *
 */
func NewTemplateValueByName(tplName string, data interface{}) TemplateValue {
	return TemplateValue{TplName: tplName, Data: data}
}

//	leafveingo 内置模板结构，用于私有的实现
type Template struct {
	funcMap       template.FuncMap   // 模板函数
	goTemplate    *template.Template // 唯一的模板管理
	isCache       bool               // 是否加入缓存 默认false
	isDevel       bool               // 是否属于开发模式 默认false
	isCompactHTML bool               //	是否压缩HTML代码格式，默认true
	leftDelims    string             //	set the action delimiters left default {{
	rightDelims   string             //	right default }}
	baseDir       string             // 模板路径的主目录 默认执行文件目录
	rwmutex       sync.RWMutex
}

//# mark Template init	----------------------------------------------------------------------------------------------------

/**
 *	init private
 */
func (l *Template) initPrivate() {
	l.funcMap = make(template.FuncMap)
	l.funcMap[kLVEmbedTempate] = l.embedTempate
	l.funcMap[kLVEmbedTempateByName] = l.embedTempateByName
	l.funcMap[kLVStringToHtml] = l.stringToHtml
	l.funcMap[kLVMapPack] = l.mapPack
	l.funcMap[kGOVersion] = runtime.Version
	l.funcMap[kLVTimeFormat] = SFTimeUtil.YMDHMSSFormat
	l.goTemplate = template.New("LVTemplate")
	l.baseDir = ""
	l.isCompactHTML = true
	l.leftDelims = "{{"
	l.rightDelims = "}}"
}

//# mark Template private method ----------------------------------------------------------------------------------------------------

/**
 *	compact html
 *
 *	@param src
 *	@param new replace string
 */
func (l *Template) compactHTML(src []byte) string {

	if l.isCompactHTML {
		// []byte{62, 60} = "><"
		src = _rexTagExtSpace.ReplaceAll(src, []byte{62, 60})
		src = _rexEmptyLineNode.ReplaceAll(src, []byte{})
		//	[]byte{36, 123, 49, 125, 32} = "${1} "
		src = _rexTagFirstSpace.ReplaceAll(src, []byte{36, 123, 49, 125, 32})
		//	[]byte{36, 123, 49, 125} = "${1}"
		src = _rexTagLastSpace.ReplaceAll(src, []byte{36, 123, 49, 125})
		src = _rexSignSpace.ReplaceAll(src, []byte{36, 123, 49, 125})
	}

	return string(src)
}

//# mark Template publc method ----------------------------------------------------------------------------------------------------

/**
 *	get template by relateve path
 *
 *	@param tplPath template relative path
 */
func (l *Template) Get(tplPath string) *template.Template {
	return l.goTemplate.Lookup(tplPath)
}

/**
 *	add cache template
 *
 *	@param name	unique name
 *	@param src	template content, if empty by name lookup cache template
 */
func (t *Template) AddCacheTemplate(tplName, src string) error {

	tmpl := t.goTemplate.New(tplName)

	if t.isCompactHTML {
		src = t.compactHTML([]byte(src))
	}

	tmpl.Delims(t.leftDelims, t.rightDelims)
	tmpl.Funcs(t.funcMap)
	_, err := tmpl.Parse(src)

	return err
}

/**
 *	pares template by relative path
 *
 *	@param tplPath
 *	@return *template.Template
 *	@return error
 */
func (l *Template) Parse(tplPath string) (tmpl *template.Template, err error) {
	//	process: golang template.New()->Funcs()->Parse()

	if l.isCache {
		if tmpl = l.goTemplate.Lookup(tplPath); nil != tmpl {
			return
		}
	}

	if fullPath := l.TemplatePathAtName(tplPath); 0 < len(fullPath) {
		var read []byte

		read, err = ioutil.ReadFile(fullPath)
		if nil != err {
			return nil, err
		}

		tplString := l.compactHTML(read)

		if l.isCache {
			tmpl = l.goTemplate.New(tplPath)
		} else {
			tmpl = template.New(tplPath)
		}

		tmpl.Delims(l.leftDelims, l.rightDelims)
		tmpl.Funcs(l.funcMap)

		if _, err = tmpl.Parse(tplString); nil != err {
			tmpl = nil
		}

	} else {
		err = errors.New("can not find :" + tplPath)
	}

	return
}

/**
 *	parse template string, not template path
 *
 *	@param name	unique name
 *	@param src	template content, if empty by name lookup cache template
 *	@return *template.Template
 *	@return error
 */
func (l *Template) ParseString(name, src string) (tmpl *template.Template, err error) {

	if l.isCache {
		if tmpl = l.goTemplate.Lookup(name); nil != tmpl {
			//	TODO 考虑如果tmpl.Parse(src)失败后，Lookup还会查询得到模版，但是永远都解析不了，这时该如何处理。
			//	查看源码Execute()时会用 escaped 参数来控制，但是是内部访问，外部访问不了，目前暂时处理不了，保留。
			return
		} else {
			tmpl = l.goTemplate.New(name)
		}
	} else {
		tmpl = template.New(name)
	}

	if l.isCompactHTML {
		src = l.compactHTML([]byte(src))
	}

	tmpl.Delims(l.leftDelims, l.rightDelims)
	tmpl.Funcs(l.funcMap)
	if _, err = tmpl.Parse(src); nil != err {
		tmpl = nil
	}

	return
}

/**
 *	execute template
 *
 *
 *	@param wr
 *	@param value
 *	@return error
 */
func (l *Template) Execute(wr io.Writer, value TemplateValue) error {
	var resultErr error
	//	TODO 待整改

	tplName := value.TplName
	tplPath := value.TplPath

	if 0 != len(tplName) {
		//	模版名称查询缓存模版
		if tmpl := l.goTemplate.Lookup(tplName); nil != tmpl {
			resultErr = tmpl.Execute(wr, value.Data)
		} else {
			resultErr = errors.New("\"" + tplName + "\" name not found template.")
		}

	} else if 0 != len(tplPath) {
		//	模版路径解析模版
		var tmpl *template.Template = nil

		tmpl, resultErr = l.Parse(tplPath)

		if nil == resultErr && nil != tmpl {
			resultErr = tmpl.Execute(wr, value.Data)
		} else if nil == resultErr {
			resultErr = errors.New(tplPath + " template parse error.")
		}

	} else {
		resultErr = errors.New("TemplateValue tplName or tplPath is nil(len == 0)")
	}

	return resultErr
}

/**
 *	template name join full path
 *
 *	@param tplName
 *	@return not find is ""
 */
func (l *Template) TemplatePathAtName(tplName string) string {
	var fullPath string

	if 0 < len(l.baseDir) {
		fullPath = path.Join(l.baseDir, tplName)
	} else {
		fullPath = path.Join(SFFileManager.GetExecDir(), tplName)
	}

	if b, isDir, _ := SFFileManager.Exists(fullPath); b && !isDir {
		return fullPath
	}

	return ""
}

/**
 *	set template func
 *
 *	default func see lv_tempate_func.go
 *		"LVEmbedTempate" LVTemplate.EmbedTempate	//嵌入模板函数
 *
 *	@param key
 *	@param methodFunc
 */
func (l *Template) SetFunc(key string, methodFunc interface{}) {
	l.rwmutex.Lock()
	defer l.rwmutex.Unlock()
	l.funcMap[key] = methodFunc
}

/**
 *	delete template func
 *
 *	@param key
 */
func (l *Template) DelFunc(key string) {
	l.rwmutex.Lock()
	defer l.rwmutex.Unlock()
	delete(l.funcMap, key)
}

/**
 *	delete all template func
 */
func (l *Template) DelAllFunc() {
	l.rwmutex.Lock()
	defer l.rwmutex.Unlock()

	for k, _ := range l.funcMap {
		delete(l.funcMap, k)
	}
}

/**
 *	set template base directory path
 *	设置模板查询主目录,默认"".
 *	主目录空的话就会按照编译文件所在目录开始查询模板
 *
 *	@param path	full path
 */
func (l *Template) SetBaseDir(path string) {
	l.baseDir = path
}

/**
 *	get template base path
 */
func (l *Template) BaseDir() string {
	return l.baseDir
}

/**
 *	set template content html format compact
 *	default true
 */
func (l *Template) SetCompactHTML(compact bool) {
	l.isCompactHTML = compact
}

/**
 *	template format compact
 */
func (l *Template) IsCompactHTML() bool {
	return l.isCompactHTML
}

/**
 *	set template is cache
 *	default false
 *
 *	@param cache
 */
func (l *Template) SetCache(cache bool) {
	l.isCache = cache
}

/**
 *	template cache
 */
func (l *Template) IsCache() bool {
	return l.isCache
}

/**
 *	is developer model
 */
func (l *Template) IsDevel() bool {
	return l.isDevel
}

/**
 *	set template developer model
 */
func (l *Template) SetDevel(isDevel bool) {
	l.isDevel = isDevel
}

/**
 *	set the action delimiters left default {{ }}
 *
 *	@param left
 *	@param right
 */
func (t *Template) SetDelims(left, right string) {
	t.leftDelims = left
	t.rightDelims = right
}
