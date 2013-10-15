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

//	leafveingo 错误信息， 使用struct进行同一管理leafveingo发生的错误
//
//	email	slowfei@foxmail.com
//	createTime 	2013-9-10
//	updateTime	2013-10-9
package leafveingo

import (
	"fmt"
)

var (
	//	http session manager 已经关闭，没有启动到session
	ErrHttpSessionManagerClosed = NewLeafveingoError("HttpSession manager has been closed.")

	ErrControllerReturnParam    = NewLeafveingoError("controllers return type parameter error.")
	ErrControllerReturnParamNum = NewLeafveingoError("controller returns the number of type parameters allowed only one.")

	//转发找不到控制器
	ErrControllerDispatcherNotFound = NewLeafveingoError("dispatcher: controller not found router key:")
)

//	leafveingo error
type LeafveingoError struct {
	Message string
}

//	new error info
func NewLeafveingoError(format string, args ...interface{}) *LeafveingoError {
	return &LeafveingoError{fmt.Sprintf(format, args...)}
}

//	error string
func (err *LeafveingoError) Error() string {
	return err.Message
}
