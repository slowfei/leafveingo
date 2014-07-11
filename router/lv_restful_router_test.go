package LVRouter

import (
	lv "github.com/slowfei/leafveingo"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type RESTfulParam struct {
	Name string
	Pwd  string
}

type TestRESTfulRouterController struct {
	tag string
}

func (c *TestRESTfulRouterController) Get(context *lv.HttpContext) interface{} {
	return "TestRESTfulRouterController"
}
func (c *TestRESTfulRouterController) Post(context *lv.HttpContext) interface{} {
	params, err := context.PackStructForm((*RESTfulParam)(nil))
	if nil != err {
		return err.Error()
	}

	paramstruct := params.(*RESTfulParam)
	return lv.BodyJson(paramstruct)
}
func (c *TestRESTfulRouterController) Put(context *lv.HttpContext) interface{} {
	session, err := context.Session(false)
	if nil != err {
		return ""
	}
	return session.UID()
}
func (c *TestRESTfulRouterController) Delete(context *lv.HttpContext) interface{} {
	return ""
}
func (c *TestRESTfulRouterController) Header(context *lv.HttpContext) interface{} {
	return ""
}
func (c *TestRESTfulRouterController) Options(context *lv.HttpContext) interface{} {
	return lv.BodyTemplateByTplName("template", "slowfei")
}

func (c *TestRESTfulRouterController) Other(context *lv.HttpContext) interface{} {
	return ""
}

//#pragma mark Test method	----------------------------------------------------------------------------------------------------
func TestRESTfulOptionsTemplate(t *testing.T) {
	req, _ := http.NewRequest("OPTIONS", "/restful/", nil)
	rw := httptest.NewRecorder()

	Server.ServeHTTP(rw, req)

	if rw.Body.String() != `<!doctype html><html lang="en"><head><meta charset="UTF-8"><title>Template</title></head><body><h1>Hello Template slowfei.</h1></body></html>` {
		t.Fatal("fatal")
	}
}
func Benchmark_TestRESTfulOptionsTemplate(b *testing.B) {
	req, _ := http.NewRequest("OPTIONS", "/restful/", nil)
	rw := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		Server.ServeHTTP(rw, req)
		rw.Body.Len()
	}
}

func TestRESTfulPostParams(t *testing.T) {
	req, _ := http.NewRequest("POST", "/restful/", nil)
	req.PostForm = make(url.Values)
	req.PostForm.Add("name", "slowfei")
	req.PostForm.Add("pwd", "slowfei-pwd")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rw := httptest.NewRecorder()

	Server.ServeHTTP(rw, req)

	if rw.Body.String() != `{"Name":"slowfei","Pwd":"slowfei-pwd"}` {
		t.Fatal("fatal")
	}
}
func Benchmark_TestRESTfulPostParams(b *testing.B) {
	req, _ := http.NewRequest("POST", "/restful/", nil)
	req.PostForm = make(url.Values)
	req.PostForm.Add("name", "slowfei")
	req.PostForm.Add("pwd", "slowfei-pwd")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rw := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		Server.ServeHTTP(rw, req)
	}
}

func TestRESTfulGetIndex(t *testing.T) {
	req, _ := http.NewRequest("GET", "/restful/", nil)
	rw := httptest.NewRecorder()

	Server.ServeHTTP(rw, req)

	if rw.Body.String() != "TestRESTfulRouterController" {
		t.Fatal("fatal")
	}
}
func Benchmark_TestRESTfulGetIndex(b *testing.B) {
	req, _ := http.NewRequest("GET", "/restful/", nil)
	rw := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		Server.ServeHTTP(rw, req)
	}
}
