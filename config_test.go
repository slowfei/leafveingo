package leafveingo

import (
	"fmt"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	c := new(Config)
	err := LoadConfigByJson([]byte(_defaultConfigJson), c)

	if nil != err {
		fmt.Println(err)
		return
	}

	fmt.Println(c.Port)
	fmt.Println(c.Addr)
	fmt.Println(c.ServerTimeout)
	fmt.Println(c.AppName)
	fmt.Println(c.AppVersion)
	fmt.Println(c.Suffixs)
	fmt.Println(c.StaticFileSuffixs)
	fmt.Println(c.Charset)
	fmt.Println(c.IsRespWriteCompress)
	fmt.Println(c.FileUploadSize)
	fmt.Println(c.IsUseSession)
	fmt.Println(c.IsGCSession)
	fmt.Println(c.SessionMaxlifeTime)
	fmt.Println(c.WebRootDir)
	fmt.Println(c.TemplateDir)
	fmt.Println(c.TemplateSuffix)
	fmt.Println(c.LogConfigPath)
	fmt.Println(c.LogChannelSize)
	c.UserData["temp"] = "tempdata"
	fmt.Println(c.UserData)
}
