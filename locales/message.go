package locales

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"
)

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
func init() {
	m := make(map[string]string)
	if err := json.Unmarshal([]byte(simChinese), &m); err != nil {
		panic(err)
	}
	locales.message["zh"] = m
}
