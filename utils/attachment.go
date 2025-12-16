package utils

import (
	"bergo/locales"
	"bytes"
)

const (
	AttachmentTypeFile = iota
	AttachmentTypeDir
)

type Attachment struct {
	Index int    `json:"index"`
	Path  string `json:"path"`
	Type  int    `json:"type"`
}

func (a *Attachment) GetContent() string {
	buf := bytes.NewBuffer(nil)
	switch a.Type {
	case AttachmentTypeDir:
		ls := LsTool{}
		dircontent := ls.List(a.Path)
		buf.WriteString(locales.Sprintf("%v. directory %v submitted as attachment:\n", a.Index, a.Path))
		buf.WriteString(dircontent)
	case AttachmentTypeFile:
		rf := ReadFile{Path: a.Path}
		filecontent, err := rf.ReadFileWhole()
		if err != nil {
			return ""
		}
		buf.WriteString(locales.Sprintf("%v. file %v submitted as attachment:\n", a.Index, a.Path))
		buf.WriteString(string(filecontent))
	}
	return buf.String()
}
