package test

import (
	"bergo/utils"
	"testing"
)

func TestLsTool(t *testing.T) {
	ls := utils.LsTool{Ig: utils.NewIgnore("..", []string{".gitignore"})}
	res := ls.ListFilesRecursive("..")
	t.Log(res)
}
