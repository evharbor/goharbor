package goharbor

import (
	"fmt"
	"testing"
)

func Init() {

}

func Test_apiBuilderStruct(t *testing.T) {
	c, err := InitConfig(map[ConfigKeyType]string{
		VERSION:   "v2",
		HOST:      "10.0.86.213:8000",
		ACCESSKEY: "666666",
		SECRETKEY: "888888",
	})
	if err != nil {
		fmt.Println(c)
	}

	builder := apiBuilderStruct{configs: c}
	params := &map[string]string{"a": "我", "b": "2"}
	url := builder.buildObjAPI("gggg", "aa", "我", params)
	if url != "" {
		fmt.Println(url)
		t.Log(url) //记录一些你期望记录的信息
	} else {
		t.Error("url build failed")
	}
}
