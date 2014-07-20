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
//  Update on 2014-07-21
//  Email  slowfei#foxmail.com
//  Home   http://www.slowfei.com

//
//	golang system some private functions implement
//
package leafveingo

import (
	"bytes"
	"crypto/tls"
	"errors"
	"net"
	"time"
)

//
// pack private conn
// implement tls http access https port some operations
//
type leafveinTCPConn struct {
	*net.TCPConn
	isTLS bool
}

/**
 *	Write
 */
func (c *leafveinTCPConn) Write(b []byte) (int, error) {

	if c.isTLS && 7 == len(b) && bytes.Equal(b, []byte{0x15, 0x3, 0x1, 0x0, 0x2, 0x2, 0xa}) {
		// error info
		// http: //golang.org/src/pkg/crypto/tls/alert.go
		// 可以尝试使用 http://localhost:443 访问tls的端口，这时浏览器不会收到任何数据。未拦截时浏览器会收到此屏蔽的二进制信息，会误以为是下载显示非常不友好。
		return 0, errors.New("unexpected message, not write.")
	}

	return c.TCPConn.Write(b)
}

//
//	reference: http://golang.org/src/pkg/net/http/server.go?h=tcpKeepAliveListener
//
type leafveinListener struct {
	*net.TCPListener
	keepAlivePeriod time.Duration
	tlsConfig       *tls.Config
}

/**
 *	Accept
 */
func (l *leafveinListener) Accept() (c net.Conn, err error) {
	tc, err := l.AcceptTCP()
	if err != nil {
		return
	}

	if 0 < l.keepAlivePeriod {
		tc.SetKeepAlive(true)
		tc.SetKeepAlivePeriod(l.keepAlivePeriod)
	}

	if nil != l.tlsConfig {
		conn := new(leafveinTCPConn)
		conn.TCPConn = tc
		conn.isTLS = true
		return tls.Server(conn, l.tlsConfig), nil
	}

	return tc, nil
}
