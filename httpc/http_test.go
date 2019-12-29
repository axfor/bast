package httpc

import "testing"

func TestHttp(t *testing.T) {
	r, err := Get("https://cn.bing.com/").Param("ccc", "vvv").String()
	if err != nil {
		t.Error(err.Error())
	} else {
		t.Log(r)
	}

	r, err = Post("https://cn.bing.com/").String()
	if err != nil {
		t.Error(err.Error())
	} else {
		t.Log(r)
	}

	err = Get("https://cn.bing.com").ToFile("./file/http_test_file.html")
	if err != nil {
		t.Error(err.Error())
	}
}
