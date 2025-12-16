package test

import (
	"bergo/utils"
	"fmt"
	"testing"
)

func TestTagFilter(t *testing.T) {
	lf := utils.NewTagFiter("think")
	content := lf.Filter("1<2 3>4 << >> <<<think>123111112212 2<3 <<>> </think> 121")
	lf.Close()
	fmt.Println(content)

	fmt.Println(lf.GetInnerConetent("think"))
}

func TestTagFilter2(t *testing.T) {
	lf := utils.NewTagFiter("think")
	tests := []string{
		"<<<thi",
		"nk>Mock thinking Mock thn2 1 > 2",
		"Mock thn3<stop></stop></think>>",
		"Mock ",
		"stream ",
		"response Bergo",
	}
	for _, test := range tests {
		content, tagContent := lf.FilterAndReturnTagContent(test, "think")
		fmt.Printf("content: %v, tagContent: %v\n", content, tagContent)
	}
}
