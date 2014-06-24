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
//  Create on 2013-9-10
//  Update on 2013-10-23
//  Email  slowfei@foxmail.com
//  Home   http://www.slowfei.com

//	leafveingo 错误信息， 使用struct进行同一管理leafveingo发生的错误
package leafveingo

import (
	"fmt"
)

var (
	//	http session manager close status
	ErrHttpSessionManagerClosed = NewLeafveinError("HttpSession manager has been closed or not started.")

	ErrControllerReturnParam    = NewLeafveinError("controllers return type parameter error.")
	ErrControllerReturnParamNum = NewLeafveinError("controller returns the number of type parameters allowed only one.")

	// 转发找不到控制器
	ErrControllerDispatcherNotFound    = NewLeafveinError("dispatcher: controller not found.")
	ErrControllerDispatcherFuncNameNil = NewLeafveinError("dispatcher: func name is nil.")

	//	Leafvein app name repeat
	ErrLeafveinAppNameRepeat = NewLeafveinError("Leafvein server app name repeat.")

	//	Leafveingo 配置对象未初始化
	ErrLeafveinInitLoadConfig    = NewLeafveinError("Leafvein initialized load config error.")
	ErrLeafveinLoadDefaultConfig = NewLeafveinError("Leafvein load default config error.")

	//	template path pase is nil
	ErrTemplatePathParseNil = NewLeafveinError("template path parse is nil.")

	//	form params
	ErrParampackParseFormParams = NewLeafveinError("form params parse error.")
	ErrParampackNewStruct       = NewLeafveinError("form params new struct error.")
)

//	leafveingo error
type LeafveinError struct {
	Message  string
	UserInfo string
}

//	new error info
func NewLeafveinError(format string, args ...interface{}) *LeafveinError {
	return &LeafveinError{fmt.Sprintf(format, args...), ""}
}

/**
 *	error string
 *
 *	@return
 */
func (err *LeafveinError) Error() string {
	errstr := "Message:(\"" + err.Message + "\");"

	if 0 != len(err.UserInfo) {
		errstr += " UserInfo:(\"" + err.UserInfo + "\")"
	}

	return errstr
}

/**
 *	equal error
 *
 *	@param
 *	@return
 */
func (err *LeafveinError) Equal(eqErr *LeafveinError) bool {
	return eqErr.Message == err.Message
}
