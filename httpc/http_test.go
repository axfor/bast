package httpc

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestHttp(t *testing.T) {
	c, err := Get("https://suggest.taobao.com/sug?code=utf-8").Param("q", "phone").String()
	if err != nil {
		t.Error(err.Error())
	}

	c2, err := Post("https://suggest.taobao.com/sug?code=utf-8&q=phone").String()
	if err != nil || c != c2 {
		t.Fatal(err)
	}

	type tb struct {
		Result [][]string `json:"result"`
	}

	rv := &tb{}
	err = Get("https://suggest.taobao.com/sug?code=utf-8&q=phone").ToJSON(rv)
	if err != nil {
		t.Fatal(err)
	}

}

func TestFile(t *testing.T) {
	f := "./file/http_test_file.html"
	err := Get("https://suggest.taobao.com/sug?code=utf-8&q=phone").ToFile(f)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll("./file")
	c, err := ioutil.ReadFile(f)
	if len(c) <= 0 || err != nil {
		t.Fatal(err)
	}
}
