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
//  Update on 2014-06-12
//  Email  slowfei@foxmail.com
//  Home   http://www.slowfei.com

//	leafveingo web http session 的封装
//
//	功能简介：
//		session高并发获取
//		自动清除session
//		session id 可选随机数生成或IPUUID
//		自动统计session生成数量、删除数量和有效数量
//		session token
//
package LVSession

import (
	// "container/list"
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"errors"
	list "github.com/slowfei/gosfcore/helper" // TODO	list 是复制新版源码的，等新版源码出来在替换
	"github.com/slowfei/gosfcore/log"
	"github.com/slowfei/gosfcore/utils/rand"
	"github.com/slowfei/gosfuuid"
	"hash"
	"net"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

//	session id 类型
type SIDType byte

const (
	SESSION_COOKIE_NAME           = "lvsessionid" //
	SESSION_COOKIE_VALUE_BASE_LEN = 22            // session cookie value 的基本长度
	COOKIE_TOKEN_KEY_RAND_LEN     = 48            // cookie token key rand bit len
	COOKIE_TOKEN_MAXLIFE_TIME     = 300           // token default maxlife time
	GLOBAL_TOKEN_KEY_LEN          = 256           // global token key rand string bit len
	FORM_TOKEN_SLICE_MAX_LEN      = 255           // form token slice storage max len，由于验证FormToken和创建时使用uint16进行转换byte,所以需要控制一定的存储大小。
	FORM_TOKEN_SLICE_DEFAULT_LEN  = 10            // form token slice default create len，由于FormToken使用于session所以在设置session所需参数的时候控制数据的一定大小，默认同一个用户请求10个页面产生10个不同的token
	DEFAULT_SCAN_GC_TIME          = 300           // 默认清理session扫描秒数,300秒
	TEMP_SESSION_MAXLIFE_TIME     = 288           // session 临时的最大有效时间，主要针对第一次请求无法获取cookie的情况下使用

	SIDTypeRandomUUID = SIDType(0) // session id 使用随机数版本类型
	SIDTypeIPUUID     = SIDType(1) // session id 使用IP版本的类型，需要链接到网络，获取失败会抛出异常

)

var (
	ErrSessionManagerFree = errors.New("session manager is freed")
	ErrCookieWrite        = errors.New("cookie can not write...")
	ErrIPValidateFail     = "ip information can not verified :%v"

	_ipFilterChar = strings.NewReplacer("[", "", "]", "")
	thisLog       = SFLog.NewLogger("LVSession")
)

//	session manager
type HttpSessionManager struct {
	CookieSecure           bool                        // cookie secure set, default false
	CookieMaxAge           int                         // cookie maxage set, default 0
	CookieTokenHash        func() hash.Hash            // cookie token hmac hash, default sha256.New
	CookieTokenRandLen     int                         // cookie token rand string len bit
	CookieTokenMaxlifeTime int32                       // cookie token maxlife time, default 300second
	globalTokenKey         []byte                      // global token hmac key, default rand string(256bit)
	scanGCTime             int64                       // 设置一个定时的时间对http session进行扫描清理,默认300秒,必须大于60秒
	listRank               map[int32]*list.List        // sessions根据有效时间排列的列队
	mapSessions            map[string]*list.Element    // 列队数列session的存储
	rwmutex                sync.RWMutex                // look
	invalidHandlerFuncs    []func(session HttpSession) // session注销前调用的函数
	sidType                SIDType                     // session id 生成类型，默认SIDTypeRandomUUID
	tempSessMaxlifeTime    int32                       // session 临时最大有效时间，默认288
	sessionCreateNum       int64                       // 当前创建session的总数量
	sessionEffectivenNum   int64                       // 当前session有效数量，没有过时的
	sessionDeleteNum       int64                       // 当前session删除的总数量
	isGC                   bool                        // 是否正在执行GC操作
	testing                bool                        // 标识测试使用
	isFree                 bool                        // 记录是否已经操作释放
}

/**
 *	new session manager
 *
 *	@param autoGC		is auto gc
 */
func NewSessionManager(autoGC bool) *HttpSessionManager {
	return NewSessionManagerAtGCTime(DEFAULT_SCAN_GC_TIME, autoGC)
}

/**
 *	new session manager
 *
 *	@param gcTimeSecond	gc operate time, min 60 second
 *	@param autoGC		is auto gc
 */
func NewSessionManagerAtGCTime(gcTimeSecond int64, autoGC bool) *HttpSessionManager {

	if 60 > gcTimeSecond {
		gcTimeSecond = 60
	}

	sm := &HttpSessionManager{mapSessions: make(map[string]*list.Element), listRank: make(map[int32]*list.List)}
	sm.scanGCTime = gcTimeSecond
	sm.sidType = SIDTypeRandomUUID
	sm.tempSessMaxlifeTime = TEMP_SESSION_MAXLIFE_TIME
	sm.testing = false
	sm.isGC = false
	sm.CookieSecure = false
	sm.CookieMaxAge = 0
	sm.CookieTokenHash = sha256.New
	sm.CookieTokenRandLen = COOKIE_TOKEN_KEY_RAND_LEN
	sm.CookieTokenMaxlifeTime = COOKIE_TOKEN_MAXLIFE_TIME

	keyBuf := make([]byte, GLOBAL_TOKEN_KEY_LEN)
	SFRandUtil.RandBits(keyBuf)
	sm.globalTokenKey = keyBuf

	if autoGC {
		go sm.autoGC()
	}

	return sm
}

/**
 *	free session manager
 *	will stop auto gc, clear all http session
 *
 */
func (sm *HttpSessionManager) Free() {
	sm.isFree = true

	//	stop auto gc

	//	remove all http session
	sessionNum := 0
	for _, lr := range sm.listRank {

		var ePTemp *list.Element = nil
		for eP := lr.Back(); nil != eP; eP = ePTemp {

			session := eP.Value.(*httpSession)
			session.sessionManager = nil

			//	如果不使用temp存储，eP被删除了还怎么指向Prev呢？
			//	所以呢，就使用一个temp进行存储Prev一个elem
			ePTemp = eP.Prev()

			//	如果maxlifeTime = -1 表示从DeleteSession删除后设置的，那时已经调用了一次函数了，所以这里不需要在调用了。
			if 0 != len(sm.invalidHandlerFuncs) && -1 != session.maxlifeTime {
				//	由于如果调用的函数占用时间久会影响GC的操作。所以copy一个session去给调用者执行其他的操作。
				//	这样做的目的也不影响GC，也不影响用户的操作，反正也是一个即将废除的session
				copySess := new(httpSession)
				*copySess = *session
				for _, f := range sm.invalidHandlerFuncs {
					go f(copySess)
				}
			}

			delete(sm.mapSessions, session.uid)
			lr.Remove(eP)

			sessionNum++
		}
	}

	if !sm.testing {
		thisLog.Info("free remove session number: %i", sessionNum)
	}
}

//	获取session，根据cookie会自动进行创建
//	每个session会根据有效时间存储在不同的列队中
//	@rw
//	@req
//	@maxlifeTime	session的最大有效时间，秒为单位。
//	@resetToken		是否重新设置cookie session token
func (sm *HttpSessionManager) GetSession(rw http.ResponseWriter, req *http.Request, maxlifeTime int32, resetToken bool) (HttpSession, error) {

	if sm.isFree {
		thisLog.Info(ErrSessionManagerFree.Error())
		return nil, ErrSessionManagerFree
	}

	var session *httpSession = nil

	cookie, _ := req.Cookie(SESSION_COOKIE_NAME)

	//	cookie大于基本的长度验证，默认uuid的base64已经是22了，加载token所以必须大于22，
	//	以后在做准确的长度效验，因为如果设置的TokenHash不同，验证的长度也会不同。
	if nil != cookie && SESSION_COOKIE_VALUE_BASE_LEN < len(cookie.Value) {
		isCreateSess := true
		signatureValue := cookie.Value

		sid, checkOk := sm.checkCookieToken(signatureValue, req.UserAgent())

		if checkOk {

			sm.rwmutex.RLock()
			element, ok := sm.mapSessions[sid]
			sm.rwmutex.RUnlock()
			isCreateSess = !ok

			if ok {

				session = element.Value.(*httpSession)

				//	IP验证，session的IP验证出现很大的分歧，考虑是否验证中。
				//	主要因为在移动设备中或一个大局域网，不是IP经常改变就是局域网IP都相同，验证反而起到不好的效果。
				//	如果不验证嘛，跨浏览器或session id 被劫持都会造成信息泄漏的可能。
				//	虽然说使用CookieSecure(https)可以有效的防止，但是非ssl呢？如何防止？不知道。
				//	目前实现记录session ip的请求规则.

				//	验证有效时间和maxlifeTime是否已经被删除了(0 >= session.maxlifeTime表示已经删除或无效了)
				if 0 >= session.maxlifeTime || (session.accessTime.Unix()+int64(session.maxlifeTime)) < time.Now().Unix() {
					isCreateSess = true
					if 0 < session.maxlifeTime {
						sm.DeleteSession(session.uid)
					}
				} else {

					//	重新设置session token
					if resetToken || (session.cookieTokenTimeunix+sm.CookieTokenMaxlifeTime) < int32(time.Now().Unix()) {
						cookie.Value = sm.generateSignature(session, req.UserAgent())
					}

					//	因为再首次设置session会使用一个临时短暂的有效时间进行设置，当第二次进行访问的时候恢复请求所设定的时间。
					//	在updateSessionRank已经做了相应的操作
					//	目前也为了统一使用锁，所以修改session交由updateSessionRank进行。
					b := sm.updateSessionRank(session, maxlifeTime)
					if !b {
						// thisLog.Info("获取session,但是已经被删除了。")
						//	能走到这部表明相同的请求再同一时间请求太多了，导致刚读取到session，然后GC就删除了。
						//	或则是调用了Invalidate将有效时间设置为-1等待GC删除，这时就需要重新创建请求了
						isCreateSess = true
					} else {

						if ip, b := sm.getUserIPAddr(req); b {
							//	记录当前访问IP并且将IP记录起来
							session.accessIP = ip
							session.recordIPAccess(ip)
						} else {
							panic(thisLog.Error(ErrIPValidateFail, req.RemoteAddr))
						}

					}

				}

			} // end ok

		} else {
			//	如果签名验证不通过有可能是人为的操作，这时可能是攻击或则其他因素，这时的创建session的有效时间就需要缩短
			maxlifeTime = sm.tempSessMaxlifeTime / 2
		}

		if isCreateSess {
			//	创建一个session
			session = sm.createSession(maxlifeTime, req)
			//	session签名
			signature := sm.generateSignature(session, req.UserAgent())
			cookie.Value = signature
		}

		//	如果signature出现修改，则需要重新设置cookie
		if signatureValue != cookie.Value {
			cookie.HttpOnly = true
			cookie.Secure = sm.CookieSecure
			cookie.Path = "/"
			http.SetCookie(rw, cookie)
			cookie.MaxAge = sm.CookieMaxAge
		}

	} else {

		//	第一次设置cookie创建一个临时的session, 最大有效时间是短暂的
		session = sm.createSession(sm.tempSessMaxlifeTime, req)

		//	session签名
		signature := sm.generateSignature(session, req.UserAgent())

		//	cookie操作，cookie放置这里是有原因的，如果放置createSession前面在并发的时候会出现一个问题，
		//	例如：两个相同的请求，都没有设置到cookie, 第一个进来，如果先添加cookie的时候，第二个也刚好进来，这时session还没有创建完毕，
		//	第二个就会利用第一个添加的cookie进入到上面判断的方法中， 然而这时sm.mapSessions[uid]是获取不到的，这样就会创建两个相同请求的
		//	session
		cookie := http.Cookie{
			Name:     SESSION_COOKIE_NAME,
			Value:    signature,
			Path:     "/",
			HttpOnly: true,
			Secure:   sm.CookieSecure,
			MaxAge:   sm.CookieMaxAge,
		}
		http.SetCookie(rw, &cookie)
		req.AddCookie(&cookie)
	}

	return session, nil
}

//	获取uuid作为session 的key
func (sm *HttpSessionManager) getUUID() (SFUUID.UUID, string) {

	var uuid SFUUID.UUID
	switch sm.sidType {
	case SIDTypeIPUUID:
		uuid = SFUUID.NewIPUUID()
	case SIDTypeRandomUUID:
		uuid = SFUUID.NewRandomUUID()
	}

	if nil == uuid {
		panic(thisLog.Error("session universally unique identifier get error."))
	}

	uuidBase64 := uuid.Base64()

	if sm.Contains(uuidBase64) {
		panic(thisLog.Error("create uuid appear repeat problem :%v", uuidBase64))
	}

	return uuid, uuidBase64
}

func (sm *HttpSessionManager) getUserIPAddr(req *http.Request) (net.IP, bool) {
	//	TODO 需要测试在使用Apache,Squid等反向代理时，需要查看IP是否获取正确，否则需要想别的办法。
	userAddr := req.RemoteAddr

	//	去除端口号
	portIndex := strings.LastIndex(userAddr, ":")
	userAddr = userAddr[:portIndex]

	ip := net.ParseIP(userAddr)

	//	如果是IP6的可能需要去除"[]"的符号。
	if nil == ip && 0 != len(userAddr) {
		userAddr = _ipFilterChar.Replace(userAddr)
		ip = net.ParseIP(userAddr)
	}

	if ip4 := ip.To4(); nil != ip4 {
		// if ip4[0] != 127 && ip4[0] != 192 {
		return ip, true
		// }
	} else if ip16 := ip.To16(); nil != ip16 {
		return ip, true
	}

	return nil, false

}

//	generate signature
//	@s session
//	@return  signature string
func (sm *HttpSessionManager) generateSignature(s *httpSession, userAgent string) (signatureBase64 string) {
	// sm.rwmutex.RLock()
	globalkey := sm.globalTokenKey
	hashFunc := sm.CookieTokenHash
	// sm.rwmutex.RUnlock()

	s.cookieTokenTimeunix = int32(time.Now().Unix())

	cookieTokenKey := make([]byte, sm.CookieTokenRandLen)
	SFRandUtil.RandBits(cookieTokenKey)

	keySHA := hashFunc()
	keySHA.Write(append(globalkey, cookieTokenKey...))
	hamcKey := keySHA.Sum(nil)

	h := hmac.New(hashFunc, hamcKey)
	h.Write(hamcKey)
	h.Write(s.uidByte)
	h.Write([]byte(userAgent))
	signature := h.Sum(nil)

	appBuf := bytes.NewBuffer(s.uidByte)
	appBuf.Write(cookieTokenKey)
	appBuf.Write(signature)
	signatureBase64 = base64.URLEncoding.EncodeToString(appBuf.Bytes())

	return
}

//	check cookie token
//  @cookieToken cookie value token
//  @userAgent
//	@return sid 	 session id
//	@return timeUnix token timeunix
//	@return ok		 pass is true
func (sm *HttpSessionManager) checkCookieToken(cookieToken, userAgent string) (sid string, ok bool) {
	ok = false

	cookieTokenByte, err := base64.URLEncoding.DecodeString(cookieToken)
	if nil != err {
		return
	}

	if len(cookieTokenByte) <= 16+sm.CookieTokenRandLen {
		return
	}

	globalkey := sm.globalTokenKey
	hashFunc := sm.CookieTokenHash
	keySHA := hashFunc()

	//	uuid 16 byte
	suid := cookieTokenByte[:16]
	cookieTokenKey := cookieTokenByte[16 : 16+sm.CookieTokenRandLen]
	signature := cookieTokenByte[len(suid)+sm.CookieTokenRandLen:]

	if len(signature) != keySHA.Size() {
		return
	}

	keySHA.Write(append(globalkey, cookieTokenKey...))
	hamcKey := keySHA.Sum(nil)

	h := hmac.New(hashFunc, hamcKey)
	h.Write(hamcKey)
	h.Write(suid)
	h.Write([]byte(userAgent))
	expectedSig := h.Sum(nil)

	ok = hmac.Equal(expectedSig, signature)
	if ok {
		sid = SFUUID.UUID(suid).Base64()
	}

	return
}

//	create session
//	@maxlifeTime
//	@request
//	@return
func (sm *HttpSessionManager) createSession(maxlifeTime int32, request *http.Request) *httpSession {
	//	由于UUID里面有函数使用了全局的锁，所以需要放置这里
	uuid, b64 := sm.getUUID()

	sm.rwmutex.Lock()
	defer sm.rwmutex.Unlock()

	s := &httpSession{}
	s.sessionManager = sm
	s.uid = b64
	s.uidByte = []byte(uuid)
	s.maxlifeTime = maxlifeTime
	s.thisMap = make(map[string]interface{})
	s.accessTime = time.Now()
	//	需要考虑存储IP的容量问题
	s.ipRule = make(map[string]int)
	s.formToken = NewFormToken(FORM_TOKEN_SLICE_DEFAULT_LEN)

	//	设置IP
	if ip, b := sm.getUserIPAddr(request); b {
		s.accessIP = ip
		s.recordIPAccess(ip)
	} else {
		panic(thisLog.Error(ErrIPValidateFail, request.RemoteAddr))
	}

	//	寻找@maxlifeTime的列队，如果寻找得到可以直接插入寻找到数列前面
	rankList, ok := sm.listRank[maxlifeTime]

	if !ok {
		rankList = list.New()
		sm.listRank[maxlifeTime] = rankList
	}

	newElement := rankList.PushFront(s)
	sm.mapSessions[s.uid] = newElement
	if !sm.testing {
		thisLog.Info("%v: create session id:%v", s.accessIP.String(), s.uid)
	}
	sm.sessionCreateNum++

	return s
}

//	更新session的列队
//	@session
//	@mltSecond	需要重置的有效访问时间
//	@return		更新成功返回true, false可能是找不到session或已经被设置最大有效时间为-1准备进行删除
func (sm *HttpSessionManager) updateSessionRank(session *httpSession, mltSecond int32) bool {
	sm.rwmutex.Lock()
	defer sm.rwmutex.Unlock()

	//	为了方便操作，session 与 sm.mapSessions[session.uid]获取的session是一致的
	element, mapOk := sm.mapSessions[session.uid]
	if !mapOk {
		//	这里如果找不到的话，很有可能是被删除了，在高并发相同的请求下就很有可能进入这里。
		//	出现这个问题基本是同一时间进行了很多个相同请求，然后又执行了自行执行销毁。
		//	如果正常请求下很少会进入这里，因为会有个有效时间。然而如果自行进行了销毁，然后该请求又再次进入，不过表示该请求已经是无作为的了。
		//	输入日志查看IP情况是否是恶意攻击。
		if !sm.testing {
			thisLog.Error("%v: update session rank, mapSessions can not find key value. uuid = %v", session.accessIP.String(), session.uid)
		}
		return false
	}

	//	发现最大有效时间小于0的话，表示要删除了，所以返回false给用户重新创建。
	if 0 >= session.maxlifeTime {
		return false
	}

	//	更换列队操作
	if session.maxlifeTime != mltSecond {
		session.maxlifeTime = mltSecond

		element.Remove()
		// if origRankL, ok := sm.listRank[session.maxlifeTime]; ok {
		// 	origRankL.Remove(element)
		// }
		if rankList, ok := sm.listRank[mltSecond]; ok {
			//	这里如果出现相同请求的高并发可能会出现重复添加的情况。
			//	不过相同请求的高并发一般是人为攻击照成的。不然高并发不同请求不会出现该问题。
			sm.mapSessions[session.uid] = rankList.PushFront(session)
		} else {
			rankList = list.New()
			sm.mapSessions[session.uid] = rankList.PushFront(session)
			sm.listRank[mltSecond] = rankList
		}

	} else {
		element.MoveToFront()
		// if origRankL, ok := sm.listRank[session.maxlifeTime]; ok {
		// 	origRankL.MoveToFront(element)
		// }
	}
	session.accessTime = time.Now()

	return true
}

func (sm *HttpSessionManager) DeleteSession(uid string) {

	if sm.isFree {
		thisLog.Info(ErrSessionManagerFree.Error())
		return
	}

	sm.rwmutex.Lock()
	defer sm.rwmutex.Unlock()

	s, ok := sm.mapSessions[uid]

	if ok {

		session := s.Value.(*httpSession)
		if 0 >= session.maxlifeTime {
			return
		}

		if 0 != len(sm.invalidHandlerFuncs) {
			//	由于如果调用的函数占用时间久会影响GC的操作。所以copy一个session去给调用者执行其他的操作。
			//	这样做的目的也不影响GC，也不影响用户的操作，反正也是一个即将废除的session
			copySess := new(httpSession)
			*copySess = *session
			for _, f := range sm.invalidHandlerFuncs {
				go f(copySess)
			}
		}
		for k, _ := range session.thisMap {
			delete(session.thisMap, k)
		}

		//	删除只能够交由GC去操作了，如果直接删除再多线程并发的时候会出现一些列无法预知的错误。
		//	直接移动到底部就OK了。
		session.maxlifeTime = -1
		s.MoveToBack()
		// sm.updateSessionRank(session, -1)

	}
}

//	设置CG HttpSession清理的间隔时间，每段时间会进行一次HttpSession的清理
//	@second	秒单位，大于或等于60m
func (sm *HttpSessionManager) SetScanGCTime(second int64) {

	if sm.isFree {
		thisLog.Info(ErrSessionManagerFree.Error())
		return
	}

	if 60 <= second {
		sm.scanGCTime = second
	}
}

//	设置临时HttpSession的最大有效时间，主要是针对第一次请求创建HttpSession所用，
//  主要为了避免垃圾HttpSession的创建，有些访问了一次并且创建了HttpSession就离
//	开了，或则Cookie无法写入时，在或者一些人的并发攻击产生大量的HttpSession所使
//	用的一个机制，在第二次访问的时候就会设置会原来的时间。
//	创建一个HttpSession最小大约占用246 bit
//
//	避免cookie无法写入时，使用一个临时短暂最大有效时间来控制session的清理。
//	再第二次请求时如果cookie能够获取得到将还原调用者设置session的最大有效时间。
//	@second 秒单位，大于或等于60m
func (sm *HttpSessionManager) SetTempSessMaxlifeTime(second int32) {

	if sm.isFree {
		thisLog.Info(ErrSessionManagerFree.Error())
		return
	}

	if 60 <= second {
		sm.tempSessMaxlifeTime = second
	}
}

//	is contains session
func (sm *HttpSessionManager) Contains(uid string) bool {

	if sm.isFree {
		thisLog.Info(ErrSessionManagerFree.Error())
		return false
	}

	sm.rwmutex.RLock()
	defer sm.rwmutex.RUnlock()
	_, ok := sm.mapSessions[uid]
	return ok
}

//	the create session total number
func (sm *HttpSessionManager) SessionCreateNum() int64 {

	if sm.isFree {
		thisLog.Info(ErrSessionManagerFree.Error())
		return -1
	}

	sm.rwmutex.RLock()
	defer sm.rwmutex.RUnlock()
	return sm.sessionCreateNum
}

//	the session effectiven number
func (sm *HttpSessionManager) SessionEffectivenNum() int64 {

	if sm.isFree {
		thisLog.Info(ErrSessionManagerFree.Error())
		return -1
	}

	sm.rwmutex.RLock()
	defer sm.rwmutex.RUnlock()
	return int64(len(sm.mapSessions))
}

// the delete session total number
func (sm *HttpSessionManager) SessionDeleteNum() int64 {

	if sm.isFree {
		thisLog.Info(ErrSessionManagerFree.Error())
		return -1
	}

	return sm.sessionDeleteNum
}

//	get cookie token key
func (sm *HttpSessionManager) GlobalTokenKey() []byte {

	if sm.isFree {
		thisLog.Info(ErrSessionManagerFree.Error())
		return nil
	}

	return sm.globalTokenKey
}

//	set cookie token key (len > 0)
//	设置全局的token key，默认是使用 crypto/rand 进行生成的随机数
func (sm *HttpSessionManager) SetGlobalTokenKey(key []byte) {

	if sm.isFree {
		thisLog.Info(ErrSessionManagerFree.Error())
		return
	}

	if 0 != len(key) {
		sm.rwmutex.Lock()
		defer sm.rwmutex.Unlock()

		sm.globalTokenKey = key
	}
}

// GC operation is running
func (sm *HttpSessionManager) IsGC() bool {
	if sm.isFree {
		thisLog.Info(ErrSessionManagerFree.Error())
		return false
	}

	sm.rwmutex.RLock()
	defer sm.rwmutex.RUnlock()
	return sm.isGC
}

//	gc clear session
func (sm *HttpSessionManager) GC() {

	if sm.isFree {
		thisLog.Info(ErrSessionManagerFree.Error())
		return
	}

	startCGTime := time.Now()
	logInfo := "http session GC start :" + startCGTime.String()
	isExecuteGC := false
	delNum := 0
	sm.rwmutex.Lock()
	sm.isGC = true
	//

	for _, lr := range sm.listRank {

		var ePTemp *list.Element = nil
		for eP := lr.Back(); nil != eP; eP = ePTemp {

			session := eP.Value.(*httpSession)
			if (session.accessTime.Unix() + int64(session.maxlifeTime)) < time.Now().Unix() {
				//	如果不使用temp存储，eP被删除了还怎么指向Prev呢？
				//	所以呢，就使用一个temp进行存储Prev一个elem
				ePTemp = eP.Prev()
				isExecuteGC = true

				session.sessionManager = nil

				//	如果maxlifeTime = -1 表示从DeleteSession删除后设置的，那时已经调用了一次函数了，所以这里不需要在调用了。
				if 0 != len(sm.invalidHandlerFuncs) && -1 != session.maxlifeTime {
					//	由于如果调用的函数占用时间久会影响GC的操作。所以copy一个session去给调用者执行其他的操作。
					//	这样做的目的也不影响GC，也不影响用户的操作，反正也是一个即将废除的session
					copySess := new(httpSession)
					*copySess = *session
					for _, f := range sm.invalidHandlerFuncs {
						go f(copySess)
					}
				}

				delete(sm.mapSessions, session.uid)
				lr.Remove(eP)

				delNum++
				sm.sessionDeleteNum++
			} else {
				break
			}

		}
	}

	if isExecuteGC {
		for k, lr := range sm.listRank {
			if 0 == lr.Len() {
				delete(sm.listRank, k)
			}
		}
		runtime.GC()
	}

	//
	sm.isGC = false
	sm.rwmutex.Unlock()
	endCGTime := time.Now()
	logInfo += "\nhttp session GC end :" + endCGTime.String()
	logInfo += "\nremoveNum:" + strconv.Itoa(delNum) + " process time:" + endCGTime.Sub(startCGTime).String() + "\n"
	thisLog.Info(logInfo)
}

//	auto gc clear
func (sm *HttpSessionManager) autoGC() {

	for {
		<-time.After(time.Duration(sm.scanGCTime) * time.Second)
		if sm.isFree {
			break
		}
		sm.GC()
	}

	// time.AfterFunc(time.Duration(sm.scanGCTime)*time.Second, sm.GC)
}

//	设置session id 的生成类型
//	@ t			SIDTypeRandomUUID or SIDTypeIPUUID
//	@ urlIPAPi	如果选择SIDTypeRandomUUID可以直接设置为 "" 空。
//				如果选择SIDTypeIPUUID，可以选择性的设置获取ip的URL地址，传递 "" 空则使用默认地址处理。
//				也可以根据自己需求，设置url, url必须能直接获取得到网络的IP信息不需要任何解析操作。
//				由于IPUUID需要连接网络，出现获取不了或解析不了IP的情况下会抛出异常(panic)
func (sm *HttpSessionManager) SetSIDType(t SIDType, urlIPApi string) {

	if sm.isFree {
		thisLog.Info(ErrSessionManagerFree.Error())
		return
	}

	switch t {
	case SIDTypeRandomUUID:
		sm.sidType = t
	case SIDTypeIPUUID:
		sm.sidType = t
		SFUUID.SetNetwordIP(urlIPApi)
	default:
		panic(thisLog.Error("session id type set error:%v", t))
	}

}

//	the invalidate session handler func
//	session will invalidate before use goroutine call
//	need to set up a function
//	parameter session is copy value
//
func (sm *HttpSessionManager) AddSessionWillInvalidateHandlerFunc(handlerFunc func(session HttpSession)) {

	if sm.isFree {
		thisLog.Info(ErrSessionManagerFree.Error())
		return
	}

	sm.invalidHandlerFuncs = append(sm.invalidHandlerFuncs, handlerFunc)
}

//
//	http session interface
type HttpSession interface {
	UID() string                                   // session id
	Get(key string) (interface{}, bool)            // get session data
	Set(key string, v interface{})                 // set session data
	Delete(key string)                             // delete
	AccessTime() time.Time                         // accessed time
	AccessIP() net.IP                              // access ip
	MaxlifeTime() int32                            // max life time
	SetMaxlifeTime(second int32)                   // set max life time
	Invalidate()                                   // invalidate this session
	IPAccessRule() map[string]int                  // record ip access rule, map[<ip string>]<ip count>
	CheckFormTokenSignature(signature string) bool // check token signature, pass return true，the token is form or query value, after checking changed token
	FormTokenSignature() string                    // form token signature, get a changed once every
}

//	http session
type httpSession struct {
	sessionManager      *HttpSessionManager    // session manager
	uid                 string                 // session id
	uidByte             []byte                 // session id uuid byte
	formToken           FormToken              // session form token, rand ( COOKIE_TOKEN_KEY_RAND_LEN bit)
	cookieTokenTimeunix int32                  // cookie session token create time unix
	thisMap             map[string]interface{} // session data
	accessTime          time.Time              // access time
	maxlifeTime         int32                  // maxlife time unit second
	accessIP            net.IP                 // access ip
	ipRule              map[string]int         // record ip access rule
	rwmutex             sync.RWMutex           // 使用锁主要防止测试或则多个相同请求并发时出现异常
}

//	记录IP访问情况
func (s *httpSession) recordIPAccess(ip net.IP) {
	//	由于防止测试的时候相同的session并发大量涌进，map未初始化不进行操作，
	//	map初始化的工作交由创建时进行
	if nil != s.ipRule {
		//	暂时注释锁的操作，如果按正常一个用户无法并发请求，除非是被攻击。所以注释留着观察。
		// s.rwmutex.Lock() // TODO
		// defer s.rwmutex.Unlock()

		key := ip.String()
		if count, ok := s.ipRule[key]; ok {
			count++
			s.ipRule[key] = count
		} else {
			s.ipRule[ip.String()] = 1
		}

	}
}

func (s *httpSession) UID() string {
	return s.uid
}

func (s *httpSession) CheckFormTokenSignature(signature string) bool {
	if 0 == len(signature) || nil == s.sessionManager {
		return false
	}
	signatureBuf, e := base64.URLEncoding.DecodeString(signature)
	if nil != e || 2 >= len(signature) {
		return false
	}

	result := false
	index := -1

	defer func() {
		//	由于已经比较过了，所以可以直接设置其他值，等调用则使用FormTokenSignature再进行重新的设置。
		// s.formToken = []byte{0x0}
		if result && 0 <= index {
			s.formToken.Remove(index)
		}
	}()

	globalkey := s.sessionManager.globalTokenKey
	hashFunc := s.sessionManager.CookieTokenHash

	indexBuf := make([]byte, 2)
	copy(indexBuf, signatureBuf[:2])

	index = int(binary.BigEndian.Uint16(indexBuf))
	randToken := s.formToken.Get(index)

	if nil == randToken || 1 >= len(randToken) || 0 > index {
		return false
	}

	shakey := hashFunc()
	shakey.Write(globalkey)
	shakey.Write(s.uidByte)
	shakey.Write(randToken)
	shakey.Write(indexBuf)
	expectedsig := shakey.Sum(indexBuf)

	// h := hmac.New(hashFunc, shakeyBuf)
	// h.Write(shakeyBuf)
	// h.Write(s.formToken)
	// expectedsig := h.Sum(nil)
	result = hmac.Equal(expectedsig, signatureBuf)
	return result
}

func (s *httpSession) FormTokenSignature() string {

	if nil == s.sessionManager {
		return "session manager is freed"
	}

	//	这里直接使用cookie token设定的长度
	randLen := s.sessionManager.CookieTokenRandLen
	globalkey := s.sessionManager.globalTokenKey
	hashFunc := s.sessionManager.CookieTokenHash

	// s.rwmutex.Lock() // TODO
	randToken := make([]byte, randLen)
	SFRandUtil.RandBits(randToken)
	index := s.formToken.Add(randToken)
	// s.rwmutex.Unlock()

	//	本来是想将index移位或计算到byte[0]头部或尾部的，然后效验时在进行获取，然后还原回原来的byte[0]，
	//	不过目前还不知道怎么去弄，不知道怎么将index隐藏到byte中。以后等有好的方案再改了。
	//	目前index使用byte进行添加到sha的头部，使用4字节。
	indexBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(indexBuf, uint16(index))

	shakey := hashFunc()
	shakey.Write(globalkey)
	shakey.Write(s.uidByte)
	shakey.Write(randToken)
	shakey.Write(indexBuf)
	signature := shakey.Sum(indexBuf)
	// h := hmac.New(hashFunc, shakeyBuf)
	// h.Write(shakeyBuf)
	// h.Write(s.formToken)
	// signature := h.Sum(nil)

	result := base64.URLEncoding.EncodeToString(signature)

	return result
}

func (s *httpSession) Get(key string) (interface{}, bool) {
	// s.rwmutex.RLock() // TODO
	// defer s.rwmutex.RUnlock()
	v, ok := s.thisMap[key]
	return v, ok
}

func (s *httpSession) Set(key string, v interface{}) {
	// s.rwmutex.Lock() // TODO
	// defer s.rwmutex.Unlock()
	s.thisMap[key] = v
}
func (s *httpSession) Delete(key string) {
	// s.rwmutex.Lock() // TODO
	// defer s.rwmutex.Unlock()

	if _, ok := s.thisMap[key]; ok {
		delete(s.thisMap, key)
	}
}

func (s *httpSession) IPAccessRule() map[string]int {
	// s.rwmutex.RLock() // TODO
	// defer s.rwmutex.RUnlock()

	newMap := make(map[string]int)
	for k, v := range s.ipRule {
		newMap[k] = v
	}
	return newMap
}

func (s *httpSession) AccessTime() time.Time {
	return s.accessTime
}
func (s *httpSession) MaxlifeTime() int32 {
	return s.maxlifeTime
}
func (s *httpSession) SetMaxlifeTime(second int32) {
	if nil == s.sessionManager {
		return
	}

	if 10 <= second {
		s.sessionManager.updateSessionRank(s, second)
	}
}
func (s *httpSession) Invalidate() {
	if nil == s.sessionManager {
		return
	}

	s.sessionManager.DeleteSession(s.uid)
}
func (s *httpSession) AccessIP() net.IP {
	return s.accessIP
}

//	formToken helper calculate storage
//	use LRU Cache mode
//	FormToken的创建原因是用于用户同时访问不同的页面，而form没有进行提交token验证，所存储的一个token
//	这样用户在打开多个设置了form token的页面时可以同时进行token的验证。
//	formToken使用于session，所以以最小的存储数据原则设定了创建的长度，使用了LRU缓存的模式，10个form token轮流使用。
//	如果用户打开了11个设置了form token的页面，第一个请求的页面验证就会失败，因为token被请求的第11个页面覆盖了。
type FormToken [][]byte

//	new formToken 指定一个长度
//	len <= 255
func NewFormToken(len int) FormToken {
	//	预留一位主要是为了添加nil的标识符号
	if FORM_TOKEN_SLICE_MAX_LEN < len {
		len = FORM_TOKEN_SLICE_MAX_LEN
	}
	return make(FormToken, len+1)
}

//	添加指定数据
//	@b	数据信息
//	@return 返回添加的下标
func (f FormToken) Add(b []byte) int {
	count := len(f)
	index := -1
	for i := 0; i < count; i++ {
		if nil == f[i] {
			f[i] = b
			f.Remove(i + 1)
			index = i
			break
		}
	}

	//	TODO 按照道理这里因该不会进来，不过如果高并发的话有可能，需要进一步进行测试。
	if -1 == index {
		thisLog.Info("FormToken Add if -1 == index")
		f[0] = b
		index = 0
	}
	return index
}

//	get the FormToken []byte
func (f FormToken) Get(i int) []byte {
	if len(f) > i {
		return f[i]
	}
	return nil
}

//	remove the FormToken []byte
func (f FormToken) Remove(i int) {
	//	为了方便Add()的操作，这里只要i大于FormToken的长度就设置0下标的位置nil，以便再次循环利用存储空间
	if len(f) > i {
		f[i] = nil
	} else {
		f[0] = nil
	}
}

//	remove all the FormToken []byte
func (f FormToken) RemoveAll() {
	count := len(f)
	for i := 0; i < count; i++ {
		f[i] = nil
	}
}
