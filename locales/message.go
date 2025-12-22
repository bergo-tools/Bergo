package locales

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// 不需要更改lang目录下的文件，这个是生成出来的。
//
//go:embed lang/zh.json
var simChinese string

type Locales struct {
	message map[string]map[string]string
	lang    string
	sync.Mutex
}

func (l *Locales) SetLang(lang string) {
	l.Lock()
	defer l.Unlock()
	l.lang = lang
}
func (l *Locales) GetTranslation(str string) string {
	l.Lock()
	defer l.Unlock()
	if _, ok := l.message[l.lang]; !ok {
		return str
	}
	if _, ok := l.message[l.lang][str]; !ok {
		return str
	}
	if l.message[l.lang][str] == "" {
		return str
	}
	return l.message[l.lang][str]
}

var locales *Locales = &Locales{
	message: make(map[string]map[string]string),
	lang:    "en",
}

func Sprintf(format string, args ...interface{}) string {
	return fmt.Sprintf(locales.GetTranslation(format), args...)
}

func Sprint(str string) {
	fmt.Println(locales.GetTranslation(str))
}

func Errorf(format string, args ...interface{}) error {
	return fmt.Errorf(locales.GetTranslation(format), args...)
}
func SetLang(lang string) {
	locales.SetLang(lang)
}
func init() {
	m := make(map[string]string)
	if err := json.Unmarshal([]byte(simChinese), &m); err != nil {
		panic(err)
	}
	locales.message["zh"] = m

	lang := os.Getenv("BERGO_LANG")
	if lang == "zh" {
		locales.SetLang("zh")
	}
}
