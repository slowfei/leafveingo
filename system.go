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
//  Create on 2014-07-15
//  Update on 2014-07-15
//  Email  slowfei#foxmail.com
//  Home   http://www.slowfei.com

//
//	golang system some private functions implement
//
package leafveingo

import (
	"net"
	"time"
)

//	reference: http://golang.org/src/pkg/net/http/server.go?h=tcpKeepAliveListener
type leafveinListener struct {
	*net.TCPListener
	keepAlivePeriod time.Duration
}

func (ln leafveinListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}

	if 0 < ln.keepAlivePeriod {
		tc.SetKeepAlive(true)
		tc.SetKeepAlivePeriod(ln.keepAlivePeriod)
	}
	return tc, nil
}
