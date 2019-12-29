package httpc

import (
	"testing"
)

func TestHttp(t *testing.T) {
	_, err := Get("https://cn.bing.com/").Param("ccc", "vvv").String()
	if err != nil {
		t.Error(err.Error())
	}

	_, err = Post("https://cn.bing.com/").String()
	if err != nil {
		t.Error(err.Error())
	}

	type tb struct {
		Result [][]string `json:"result"`
	}

	rv := &tb{}
	err = Get("https://suggest.taobao.com/sug?code=utf-8&q=phone").ToJSON(rv)
	if err != nil {
		t.Error(err.Error())
	}

}

func TestFile(t *testing.T) {
	err := Get("https://cn.bing.com").ToFile("./file/http_test_file.html")
	if err != nil {
		t.Error(err.Error())
	}
}
