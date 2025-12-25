package test

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
)

func TestAutoStyleFunc(t *testing.T) {
	tr, _ := glamour.NewTermRenderer(glamour.WithAutoStyle())
	val := reflect.ValueOf(tr).Elem().FieldByName("ansiOptions")
	//mock := ansi.StyleConfig{}
	stylePtt := (*ansi.Options)(unsafe.Pointer(val.UnsafeAddr()))

	t.Logf("%v %v", stylePtt, val.CanSet())
}
