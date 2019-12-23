package lang

import (
	"strconv"
	"strings"
)

var trans = map[string][]TranItem{}
var keyTrans = map[string]string{}

//TranItem a trans item
type TranItem struct {
	Text  string
	Index int
}

//Trans translator
func Trans(lang, key string, param ...string) string {
	if lang == "" {
		lang = "en"
	}
	key = lang + "." + key
	if t, ok := trans[key]; ok {
		lg := 0
		if param != nil {
			lg = len(param)
		}
		var s strings.Builder
		for _, v := range t {
			if v.Index >= 0 && v.Index < lg {
				s.WriteString(param[v.Index])
			} else {
				s.WriteString(v.Text)
			}
		}
		return s.String()
	}
	return ""
}

//Key translator key
func Key(lang, key string) string {
	if lang == "" {
		lang = "en"
	}
	if v, ok := keyTrans[lang+"."+key]; ok {
		return v
	}
	return key
}

//Register a translator provide by the trans name
func Register(lang string, ts map[string]string) {
	for k, v := range ts {
		vs := lang + "." + k
		if _, ok := trans[vs]; !ok {
			vv := v
			lg := len(v)
			trs := []TranItem{}
			for {
				i := strings.Index(vv, "{")
				j := strings.Index(vv, "}")
				if i == -1 || j == -1 || i >= j {
					break
				}
				if i > 0 {
					trs = append(trs, TranItem{vv[0:i], -1})
				}
				i++
				p := vv[i:j]
				pi, err := strconv.Atoi(p)
				if err != nil {
					break
				}
				rp := "{" + p + "}"
				trs = append(trs, TranItem{rp, pi})
				j++
				if j < lg {
					vv = vv[j:]
					continue
				}
				vv = ""
				break
			}
			if vv != "" {
				trs = append(trs, TranItem{vv, -1})
			}
			trans[vs] = trs
		}
	}
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
