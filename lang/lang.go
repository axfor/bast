//Copyright 2018 The axx Authors. All rights reserved.

package lang

import (
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

var trans = map[string]*translator{}

var keyTrans = map[string]string{}

type translator struct {
	Item []*item
	Cap  int
}

type item struct {
	Text  string
	Index int
}

//Trans translator
func Trans(lang, key string, param ...string) string {
	if lang == "" {
		lang = "en"
	}
	k := lang + "." + key
	if t, ok := trans[k]; ok {
		lg := 0
		if param != nil {
			lg = len(param)
		}
		var s strings.Builder
		s.Grow(t.Cap)
		for _, v := range t.Item {
			if v.Index >= 0 && v.Index < lg {
				s.WriteString(param[v.Index])
			} else {
				s.WriteString(v.Text)
			}
		}
		return s.String()
	}
	return key
}

//Transk translator of key
func Transk(lang, key string) string {
	if lang == "" {
		lang = "en"
	}
	if v, ok := keyTrans[lang+"."+key]; ok {
		return v
	}
	return key
}

//Register a translator provide by the trans name
func Register(lang string, ts map[string]string) error {
	for k, v := range ts {
		vs := lang + "." + k
		//if _, ok := trans[vs]; !ok {
		vv := v
		lg := len(v)
		trs := []*item{}
		refCap := lg
		for {
			i := strings.Index(vv, "{")
			j := strings.Index(vv, "}")
			if i == -1 || j == -1 || i >= j {
				break
			}
			if i > 0 {
				trs = append(trs, &item{vv[0:i], -1})
			}
			i++
			p := vv[i:j]
			pi, err := strconv.Atoi(p)
			if err != nil {
				break
			}
			rp := "{" + p + "}"
			trs = append(trs, &item{rp, pi})
			j++
			if j < lg {
				vv = vv[j:]
				continue
			}
			vv = ""
			break
		}
		if vv != "" {
			trs = append(trs, &item{vv, -1})
		}
		lg = len(trs)
		if lg > 0 {
			refCap += lg * 20
			trans[vs] = &translator{
				Item: trs,
				Cap:  refCap,
			}
		}
		//}
	}
	return nil
}

//File translator file
func File(file string) error {
	if file == "" {
		return nil
	}
	lang := "en"
	fn := path.Base(file)
	i := strings.LastIndex(fn, ".")
	if i > 0 {
		j := strings.LastIndex(fn[0:i], ".")
		if j >= 0 {
			j++
			lang = fn[j:i]
		} else {
			lang = fn[0:i]
		}
	}
	if lang == "" {
		lang = "en"
	}
	data, err := ioutil.ReadFile(file)
	if len(data) <= 0 || err != nil {
		return err
	}
	t := map[string]string{}
	err = yaml.Unmarshal([]byte(data), &t)
	if err != nil {
		return err
	}
	return Register(lang, t)
}

//Dir translator dir
func Dir(dir string) error {
	if dir == "" {
		return nil
	}
	f, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !f.IsDir() {
		return File(dir)
	}
	fs, err := ioutil.ReadDir(dir)
	if len(fs) <= 0 || err != nil {
		return err
	}
	for _, f := range fs {
		if !f.IsDir() {
			err = File(dir + "/" + f.Name())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

//TransFile translator dir or file
func TransFile(f string) error {
	if f == "" {
		return nil
	}
	return Dir(f)
}

//RegisterKeys a translator provide by the key trans name
func RegisterKeys(lang string, ks map[string]string) {
	for k, v := range ks {
		vs := lang + "." + k
		if _, ok := keyTrans[vs]; !ok {
			keyTrans[vs] = v + ""
		}
	}
}

//RegisterKey a translator provide by the key tran name
func RegisterKey(lang string, key, tran string) {
	vs := lang + "." + key
	if _, ok := keyTrans[vs]; !ok {
		keyTrans[vs] = tran
	}
}
