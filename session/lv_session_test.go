package LVSession

import (
	// "crypto/sha1"
	"fmt"
	"github.com/slowfei/gosfcore/utils/rand"
	"net/http"
	"net/url"
	"runtime"
	// "strings"
	"sync"
	"testing"
	"time"
)

type TestImplResponse struct {
}

func (t TestImplResponse) Header() http.Header {
	return make(http.Header, 20)
}
func (t TestImplResponse) Write([]byte) (int, error) {
	return 1, nil
}
func (t TestImplResponse) WriteHeader(i int) {
}

/**
 *	测试session manager的Free()
 *	测试时间20秒，需要耐心等待
 */
func TestSessionManagerFree(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	sm := NewSessionManagerAtGCTime(1, true)

	sm.testing = true

	fmt.Println("sm free() testing...:", time.Now(), "\n")

	go func() {

		//	添加http session
		for i := 1; i <= 10; i++ {
			rw := TestImplResponse{}
			req := http.Request{}
			req.RemoteAddr = "128.0.0.1:8212"
			req.Header = make(http.Header, 1)
			req.Form = make(url.Values)
			sm.tempSessMaxlifeTime = int32(i)
			_, err := sm.GetSession(rw, &req, int32(i), false)
			if nil != err {
				fmt.Println("get session error:", err)
			}
		}

		<-time.After(time.Duration(6) * time.Second)
		fmt.Println("session manager start free()...:", time.Now())
		sm.Free()
	}()

	time.Sleep(time.Duration(20) * time.Second)
}

/**
 *	测试cookie 签名效验，这里操作修改特定的token字符，看看是否可以验证通过
 */
func TestCookieReplaceChar(t *testing.T) {
	sm := NewSessionManagerAtGCTime(10, true)
	// sm.CookieTokenHash = sha1.New
	// rw := TestImplResponse{}
	req := http.Request{}
	req.RemoteAddr = "128.0.0.1:8212"
	req.Header = make(http.Header, 1)

	s := &httpSession{}
	uuid, b64 := sm.getUUID()
	s.uid = b64
	s.uidByte = []byte(uuid)
	s.maxlifeTime = 10
	s.thisMap = make(map[string]interface{})
	s.accessTime = time.Now()
	s.ipRule = make(map[string]int)
	s.formToken = NewFormToken(10)

	startTime := time.Now()
	signature := sm.generateSignature(s, "slowfei---")
	fmt.Println("time:", time.Now().Sub(startTime))
	fmt.Println("signature:", signature)

	fmt.Println("测试修改每一个字符...")
	for i := 0; i < len(signature); i++ {
		repByte := []byte(signature)
		if repByte[i] != byte('B') {
			repByte[i] = byte('B')
		} else {
			repByte[i] = 0x33
		}

		newSignature := string(repByte)
		_, ok := sm.checkCookieToken(newSignature, "slowfei---")
		if ok {
			t.Fatal("发现经过修改的字符验证能够通过，这是一个严重错误。index:", i)
		}
	}
	fmt.Println("测试通过。")
}

/**
 *	测试cookie token的验证操作
 */
func TestCookieToken(t *testing.T) {
	sm := NewSessionManagerAtGCTime(10, true)
	// sm.CookieTokenHash = sha1.New
	// rw := TestImplResponse{}
	req := http.Request{}
	req.RemoteAddr = "128.0.0.1:8212"
	req.Header = make(http.Header, 1)

	s := &httpSession{}
	uuid, b64 := sm.getUUID()
	s.uid = b64
	s.uidByte = []byte(uuid)
	s.maxlifeTime = 10
	s.thisMap = make(map[string]interface{})
	s.accessTime = time.Now()
	s.ipRule = make(map[string]int)
	s.formToken = NewFormToken(10)

	//	判断创建时间和验证的时间各是多少。
	startTime := time.Now()
	signature := sm.generateSignature(s, "slowfei---")
	fmt.Println("time:", time.Now().Sub(startTime))
	fmt.Println("signature:", signature)
	fmt.Println(len(signature))
	fmt.Println("")

	//	这里尝试替换一个字符
	// signature = strings.Replace(signature, "D", "d", 1)
	// fmt.Println("signature:", signature)
	// repByte := []byte(signature)
	// repByte[127] = 0x32
	// signature = string(repByte)
	// fmt.Println("signature:", signature)
	startTime = time.Now()
	sid, ok := sm.checkCookieToken(signature, "slowfei---")
	fmt.Println("time:", time.Now().Sub(startTime))
	fmt.Println(sid, ok)

	if !ok {
		t.Fatal("checkCookieToken fatal")
		return
	}

	if sid != b64 {
		t.Fatalf("check returns sid != create sid:\nreturn sid:%v\nb64:", sid, b64)
	}

	fmt.Println("")
	fmt.Println("测试FormToken...")
	startTime = time.Now()
	formToken := s.FormTokenSignature()
	fmt.Println("time:", time.Now().Sub(startTime))
	fmt.Println("formToken:", formToken)

	//	这里尝试替换一个字符
	// formToken = strings.Replace(formToken, "D", "d", 1)
	// fmt.Println("formToken:", formToken)
	startTime = time.Now()
	ok = s.CheckFormTokenSignature(formToken)
	fmt.Println("time:", time.Now().Sub(startTime))
	fmt.Println("check:", ok)
	if !ok {
		t.Fatal("CheckFormTokenSignature fatal")
		return
	}

	// return

	fmt.Println("")
	fmt.Println("测试并发FormToken...")
	runtime.GOMAXPROCS(runtime.NumCPU())
	forNum := 10 //	进行多少次并发，这里不要随意更改测试数量，少点较号，因为同一个session同时访问不可能回出现很多。
	goNum := 10  //	每次并发多少
	wgNum := goNum * forNum
	wg := sync.WaitGroup{}
	wg.Add(wgNum)
	var rwm sync.RWMutex

	//	测试时注意FormeToken的长度，最长设置为255
	s.formToken = NewFormToken(wgNum / 2)
	formTokens := make([]string, wgNum)

	for i := 0; i < forNum; i++ {
		for j := 0; j < goNum; j++ {
			ftIndex := i*goNum + j
			go func() {
				rwm.Lock()
				ft := s.FormTokenSignature()
				formTokens[ftIndex] = ft
				rwm.Unlock()
				// fmt.Println("index: ", ftIndex, " -- ", ft)
				if ftIndex%10 == 0 {
					rwm.Lock()

					ok := s.CheckFormTokenSignature(ft)
					if ok {
						formTokens[ftIndex] = ""
					} else {
						fmt.Println("线程并发中验证时Check失败：", ft)
					}

					rwm.Unlock()
				}
				wg.Done()
			}()
		}
	}
	wg.Wait()

	checkFalseNum := 0
	checkTrueNum := 0
	fmt.Println("观察下formToken的byte存储顺序：")
	//	验证formToken的存储数据，肯定有一个会是nil，因为再设置长度的时候讲长度+1，主要是使用了LRU Cache mode，循环进行存储。
	//	并发测试的目的是保证在多并发的时候不出现其他的系统级或则代码处理的异常错误
	//	CheckFormTokenSignature验证失败可能是token被覆盖了，这个不要什么紧。主要在测试页面的时候提交请求能够正常验证就好。
	//	暂时取消打印
	// for i, v := range s.formToken {
	// 	fmt.Println(i, " byte nil=", v == nil)
	// }

	for _, v := range formTokens {
		if v != "" {
			ok := s.CheckFormTokenSignature(v)
			if !ok {
				//	这里出现错误不奇怪，因为使用的时并发测试，有可能存储的formToken的下标序号被覆盖了。
				// fmt.Printf("go func test form token: index:%v (len:%v)%v \n", i, len(v), v)
				checkFalseNum++
			} else {
				checkTrueNum++
			}
		}
	}
	fmt.Println("验证成功数:", checkTrueNum, "  失败数:", checkFalseNum)
	fmt.Println("并发FormToken测试完成")

}

/**
 *	测试GetSession的并发情况
 */
func TestRquestGetSession(t *testing.T) {

	runtime.GOMAXPROCS(runtime.NumCPU())
	sm := NewSessionManagerAtGCTime(1, true)
	sm.tempSessMaxlifeTime = 5
	sm.testing = true

	// 统计记录
	errorNum := 0                         //	创建错误session统计
	delSessCount := 0                     //	删除session统计
	sessCreateMap := make(map[string]int) //	创建session统计
	getSessNum := 0                       //	获取session的次数
	isResetToken := true                  //	每次请求重置session token
	// sameReqMap := make(map[string]int)    //	相同请求的计算

	//	并发数量别乱调，小心系统崩溃
	var maxlifeTime int32 = 10 //	session最大有效时间
	forNum := 1000             //	进行多少次并发
	goNum := 100               //	每次并发多少
	wgNum := goNum * forNum
	wg := sync.WaitGroup{}
	wg.Add(wgNum)
	var rwm sync.RWMutex

	sm.AddSessionWillInvalidateHandlerFunc(func(session HttpSession) {
		if nil == session {
			fmt.Println("InvalidateHandlerFunc session nil....")
		}
		rwm.Lock()
		delSessCount++
		rwm.Unlock()
		time.Sleep(time.Duration(1) * time.Second)
	})

	//	request rw num
	reqCount := wgNum / 5
	rws := make([]TestImplResponse, reqCount)
	reqs := make([]*http.Request, reqCount)
	for i := 0; i < reqCount; i++ {
		rws[i] = TestImplResponse{}

		req := http.Request{}
		req.RemoteAddr = "128.0.0.1:8212"
		req.Header = make(http.Header, 1)
		req.Form = make(url.Values)
		reqs[i] = &req
	}

	for i := 0; i < forNum; i++ {
		for j := 0; j < goNum; j++ {
			// fmt.Println(i*forNum + j)
			forIndex := i*goNum + j
			go func() {
				rw := TestImplResponse{}
				req := &http.Request{}
				req.RemoteAddr = "128.0.0.1:8212"
				req.Header = make(http.Header, 1)
				req.Form = make(url.Values)

				sameCookieV := ""
				if forIndex%5 == 0 {
					//	随机获取相同的请求
					rint := SFRandUtil.RandBetweenInt(0, int64(reqCount)-1)
					rw = rws[rint]
					req = reqs[rint]

					//	相同请求计算起来特别累，不知道怎么计算，而且好像计算也没用，这个是客户端的问题，只要服务端session能够正常GC就好。
					// ce, _ := req.Cookie(SESSION_COOKIE_NAME)
					// if nil != ce && "" != ce.Value {
					// 	rwm.Lock()
					// 	if count, ok := sameReqMap[ce.Value]; ok {
					// 		sameReqMap[ce.Value] = count + 1
					// 	} else {
					// 		sameReqMap[ce.Value] = 1
					// 	}
					// 	sameCookieV = ce.Value
					// 	rwm.Unlock()
					// }
				}

				session, err := sm.GetSession(rw, req, maxlifeTime, isResetToken)

				rwm.Lock()
				getSessNum++
				rwm.Unlock()

				if nil != err {
					rwm.Lock()
					errorNum++
					rwm.Unlock()
				} else {
					rwm.Lock()
					if count, ok := sessCreateMap[session.UID()]; ok {
						sessCreateMap[session.UID()] = count + 1
					} else {
						sessCreateMap[session.UID()] = 1
					}
					if "" != sameCookieV {
						//	可以观察session id 与 cookie是否是相同的，看cookie前几位就知道了，前几位是UUID的存储
						//	如果出现不相同可能要检查下cookie token的验证，可能导致了创建了一个新的session
						// fmt.Println(sameCookieV)
						// fmt.Println(session.UID())
					}

					rwm.Unlock()

					if forIndex%15 == 0 {
						session.Invalidate()
					}
				}

				wg.Done()
			}()
		}
	}

	wg.Wait()
	time.Sleep(time.Duration(30) * time.Second)

	//	等待GC最后一次的执行
	if sm.IsGC() {
		for {
			if !sm.IsGC() {
				break
			}
		}
	}

	//	测试要求的结果，只要运行起来不出现系统级别或代码错误(panic)就根据以下判断的结果进行比较就可以了。
	//	创建session数量与GetSession不相同可能是相同的请求照成的，不要紧的。

	//	计算SessionManager存储的session是否正确
	mapSessNum := len(sm.mapSessions)
	listRankSessNum := 0
	for _, v := range sm.listRank {
		listRankSessNum += v.Len()
	}
	fmt.Println("mapSessions num:", mapSessNum)
	fmt.Println("listRankSession num:", listRankSessNum)
	if mapSessNum != listRankSessNum {
		//	如果出现两个不相等就表示存储出逻辑问题，需要调查session的存储情况
		t.Fatalf("mapSessNum != listRankSessNum")
		return
	}

	//	计算相同的请求
	//	相同请求计算起来特别累，不知道怎么计算，而且好像计算也没用，这个是客户端的问题，只要服务端session能够正常GC就好。
	// sameReqNum := 0
	// for _, v := range sameReqMap {
	// 	sameReqNum = sameReqNum + v
	// }
	// fmt.Println("same request num:", sameReqNum)

	//	进行了多少次GetSession
	fmt.Println("GetSession(...) num:", getSessNum)
	fmt.Println("GetSession(...) Error num:", errorNum)

	//	计算创建的session，创建数量与删除数是一致的，如果出现不一致则需要进一步测试，不一致肯定是有错误的。
	sessionCreateNum := len(sessCreateMap)
	fmt.Println("create session num:", sessionCreateNum)
	fmt.Println("delete session num:", delSessCount)

	fmt.Println("not delete session num:", mapSessNum)
	if delSessCount+mapSessNum != sessionCreateNum {
		//	如果这里出现错误表示GC删除的结果与创建session的结果不一致，这里需要调查GC删除是否有错误没有计算到
		t.Fatal("delSessCount+mapSessNum != sessionCreateNum")
		return
	}
	// for k, v := range sessCreateMap {

	// }

}
