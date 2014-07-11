package LVRouter

import (
	// "fmt"
	lv "github.com/slowfei/leafveingo"
	lvs "github.com/slowfei/leafveingo/session"
	"net/http"
	"net/http/httptest"
	// "runtime"
	"net/url"
	"testing"
)

//
//	reflect controller
//
type TestReflectController struct {
	tag string
}

func (rc *TestReflectController) Index() interface{} {
	return "TestReflectController"
}

func (rc *TestReflectController) Params(params struct {
	Name string
	Pwd  string
}) interface{} {
	return lv.BodyJson(params)
}
func (rc *TestReflectController) PostParams(params struct {
	Name string
	Pwd  string
}) interface{} {
	return lv.BodyJson(params)
}

func (rc *TestReflectController) Session(session lvs.HttpSession) interface{} {
	return session.UID()
}
func (rc *TestReflectController) SessionParams(session lvs.HttpSession, params struct {
	Name string
	Pwd  string
}) interface{} {

	if 0 == len(params.Name) || 0 == len(params.Pwd) {
		return ""
	}

	return session.UID()
}

func (rc *TestReflectController) Template() interface{} {
	return lv.BodyTemplateByTplName("template", "slowfei")
}

//#pragma mark Test method	----------------------------------------------------------------------------------------------------

func TestReflectTemplate(t *testing.T) {
	req, _ := http.NewRequest("GET", "/template", nil)
	rw := httptest.NewRecorder()

	Server.ServeHTTP(rw, req)

	if rw.Body.String() != `<!doctype html><html lang="en"><head><meta charset="UTF-8"><title>Template</title></head><body><h1>Hello Template slowfei.</h1></body></html>` {
		t.Fatal("fatal")
	}
}

func Benchmark_TestReflectTemplate(b *testing.B) {
	req, _ := http.NewRequest("GET", "/template", nil)
	rw := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		Server.ServeHTTP(rw, req)
		rw.Body.Len()
	}
}

func TestReflectSession(t *testing.T) {
	req, _ := http.NewRequest("GET", "/session", nil)
	req.RemoteAddr = "127.0.0.1"
	rw := httptest.NewRecorder()

	Server.ServeHTTP(rw, req)

	if 0 == len(rw.Body.String()) {
		t.Fatal("fatal")
	}

	req, _ = http.NewRequest("GET", "/session/params?name=slowfei&pwd=slowfei-pwd", nil)
	req.RemoteAddr = "127.0.0.1"
	rw = httptest.NewRecorder()

	Server.ServeHTTP(rw, req)

	if 0 == len(rw.Body.String()) || 200 != rw.Code {
		t.Fatal("fatal")
	}

}

func Benchmark_TestReflectSession(b *testing.B) {
	req, _ := http.NewRequest("GET", "/session", nil)
	req.RemoteAddr = "127.0.0.1"
	rw := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		Server.ServeHTTP(rw, req)
	}
}
func Benchmark_TestReflectSessionParams(b *testing.B) {
	req, _ := http.NewRequest("GET", "/session/params?name=slowfei&pwd=slowfei-pwd", nil)
	req.RemoteAddr = "127.0.0.1"
	rw := httptest.NewRecorder()

	pass := false

	for i := 0; i < b.N; i++ {
		Server.ServeHTTP(rw, req)

		if !pass {
			if 0 == len(rw.Body.String()) || 200 != rw.Code {
				b.Fatal("fatal")
			} else {
				pass = true
			}
		}
	}
}

func TestReflectParams(t *testing.T) {
	req, _ := http.NewRequest("GET", "/params?name=slowfei&pwd=slowfei-pwd", nil)
	rw := httptest.NewRecorder()

	Server.ServeHTTP(rw, req)

	if rw.Body.String() != `{"Name":"slowfei","Pwd":"slowfei-pwd"}` {
		t.Fatal("fatal")
	}

	req, _ = http.NewRequest("POST", "/params", nil)
	req.PostForm = make(url.Values)
	req.PostForm.Add("name", "slowfei")
	req.PostForm.Add("pwd", "slowfei-pwd")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rw = httptest.NewRecorder()

	Server.ServeHTTP(rw, req)

	if rw.Body.String() != `{"Name":"slowfei","Pwd":"slowfei-pwd"}` {
		t.Fatal("fatal")
	}
}

func Benchmark_TestReflectParams(b *testing.B) {
	req, _ := http.NewRequest("GET", "/params?name=slowfei&pwd=slowfei-pwd", nil)
	rw := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		Server.ServeHTTP(rw, req)
	}
}

func Benchmark_TestReflectParams_POST(b *testing.B) {
	req, _ := http.NewRequest("POST", "/params", nil)
	req.PostForm = make(url.Values)
	req.PostForm.Add("name", "slowfei")
	req.PostForm.Add("pwd", "slowfei-pwd")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rw := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		Server.ServeHTTP(rw, req)
	}
}

func TestReflectIndex(t *testing.T) {

	req, _ := http.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()

	Server.ServeHTTP(rw, req)

	if rw.Body.String() == "TestReflectController1" {
		t.Fatal("fatal")
	}
}

func Benchmark_TestReflectIndex(b *testing.B) {
	req, _ := http.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {

		Server.ServeHTTP(rw, req)
	}
}
