package utils

import (
	"bytes"
	"fmt"
	"strings"
)

// 过滤 XML 标签里的内容出来
const (
	stateText = iota
	stateSuspectTag
	stateInTag
	stateEndTag
)

type TagContentItem struct {
	Tag          string `json:"tag"`           //标签名称
	InnerContent string `json:"inner_content"` //标签内的内容
	WholeContent string `json:"whole_content"` //标签完整内容
	err          bool
}
type TagFiter struct {
	TagContent    []*TagContentItem
	registerTag   map[string]bool
	finishHooks   map[string]func(string)
	beginHooks    map[string]func(string)
	state         int
	activeTag     string
	tagBuf        *bytes.Buffer
	wholeStartTag string
}

func NewTagContent(inner string, tag string, errs ...bool) *TagContentItem {
	err := false
	if len(errs) > 0 {
		err = true
	}
	return &TagContentItem{
		Tag:          tag,
		InnerContent: inner,
		WholeContent: fmt.Sprintf("<%s>%s</%s>", tag, inner, tag),
		err:          err,
	}
}
func NewTagFiter(tags ...string) *TagFiter {
	lf := &TagFiter{
		TagContent:  make([]*TagContentItem, 0),
		state:       stateText,
		tagBuf:      bytes.NewBuffer(nil),
		finishHooks: make(map[string]func(string)),
		registerTag: make(map[string]bool),
		beginHooks:  make(map[string]func(string)),
	}
	for _, tag := range tags {
		lf.registerTag[tag] = true
	}
	return lf
}

func (lf *TagFiter) AddFinishHook(tag string, hook func(string)) {
	lf.registerTag[tag] = true
	lf.finishHooks[tag] = hook
}

func (lf *TagFiter) AddBeginHook(tag string, hook func(string)) {
	lf.registerTag[tag] = true
	lf.beginHooks[tag] = hook
}

func (lf *TagFiter) Filter(content string) string {
	output := bytes.NewBuffer(nil)
	for _, r := range content {
		output.WriteString(lf.FilterRune(r))
	}
	return output.String()
}

// tag中的内容放在第二个返回值中
func (lf *TagFiter) FilterAndReturnTagContent(content string, tag string) (string, string) {
	output := bytes.NewBuffer(nil)
	tagOutput := bytes.NewBuffer(nil)
	for _, r := range content {
		state := lf.state
		filtered := lf.FilterRune(r)
		if state == stateSuspectTag && lf.state == stateInTag && tag == lf.activeTag {
			tagOutput.WriteString(lf.tagBuf.String())
		}
		if state == stateInTag && tag == lf.activeTag {
			tagOutput.WriteString(string(r))
		}
		output.WriteString(filtered)
	}
	return output.String(), tagOutput.String()
}

func (lf *TagFiter) FilterRune(r rune) string {
	switch lf.state {
	case stateText:
		lf.tagBuf.WriteRune(r)
		if r == '<' {
			lf.state = stateSuspectTag
		} else {
			return lf.flushTag()
		}
	case stateSuspectTag:
		if r == '>' {
			lf.tagBuf.WriteRune(r)
			wholeTag := lf.tagBuf.String()
			tag := lf.extracTag(wholeTag)
			if tag == "" {
				lf.state = stateText
				return lf.flushTag()
			} else if _, exists := lf.registerTag[tag]; !exists {
				lf.state = stateText
				return lf.flushTag()
			} else {
				lf.state = stateInTag
				lf.activeTag = tag
				lf.wholeStartTag = wholeTag
				if _, exists := lf.beginHooks[tag]; exists {
					lf.beginHooks[tag](lf.tagBuf.String())
				}
			}
		} else if r == '<' {
			res := lf.flushTag()
			lf.tagBuf.WriteRune(r)
			return res
		} else {
			lf.tagBuf.WriteRune(r)
		}
	case stateInTag:
		lf.tagBuf.WriteRune(r)
		if r == '>' {
			if strings.HasSuffix(lf.tagBuf.String(), "</"+lf.activeTag+">") {
				lf.state = stateText
				wholeConent := lf.tagBuf.String()
				item := NewTagContent(lf.getInner(wholeConent), lf.activeTag)
				lf.TagContent = append(lf.TagContent, item)
				if _, exists := lf.finishHooks[lf.activeTag]; exists {
					lf.finishHooks[lf.activeTag](item.InnerContent)
				}
				lf.tagBuf.Reset()
			}
		}
	}
	return ""
}

func (lf *TagFiter) GetInnerConetent(tag string) string {
	for _, item := range lf.TagContent {
		if item.Tag == tag {
			return item.InnerContent
		}
	}
	return ""
}

func (lf *TagFiter) GetWholeConetent(tag string) string {
	for _, item := range lf.TagContent {
		if item.Tag == tag {
			return item.WholeContent
		}
	}
	return ""
}

func (lf *TagFiter) flushTag() string {
	tmp := lf.tagBuf.String()
	lf.tagBuf.Reset()
	return tmp
}

func (lf *TagFiter) extracTag(label string) string {
	if len(label) < 2 {
		return ""
	}
	if !(strings.HasPrefix(label, "<") && strings.HasSuffix(label, ">")) {
		return ""
	}
	buf := bytes.NewBuffer(nil)
	for _, r := range label {
		if r == '<' || r == '>' {
			continue
		}
		if r == ' ' || r == '\n' {
			break
		}
		buf.WriteRune(r)

	}
	return buf.String()
}
func (lf *TagFiter) getInner(cont string) string {
	res := strings.TrimPrefix(cont, lf.wholeStartTag)
	res = strings.TrimSuffix(res, "</"+lf.activeTag+">")
	return res
}
func (lf *TagFiter) Close() (string, error) {
	if lf.state == stateInTag {
		return lf.flushTag(), fmt.Errorf("unclosed tag <%s>", lf.activeTag)
	}
	return lf.flushTag(), nil
}
