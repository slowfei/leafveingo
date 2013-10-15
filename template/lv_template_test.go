package LVTemplate

import (
	"fmt"
	"os"
	"strconv"
	"testing"
)

func AddTest(a, b int) string {
	return strconv.Itoa(a + b)
}

func PrintData(data interface{}) string {
	fmt.Println("PrintData:", data)
	return ""
}

func TestExecute(t *testing.T) {
	tv := NewTemplateValue("hellp.tpl", map[string]string{"Name": "slowfei", "Tag": "sf_tag", "Title": "标题"})
	lvTempl := SharedTemplate()
	lvTempl.SetBaseDir("")
	lvTempl.AddFunc("AddTest", AddTest)
	lvTempl.AddFunc("PrintData", PrintData)
	err := lvTempl.Execute(tv, os.Stdout)
	if nil != err {
		fmt.Println(err)
	}

}
