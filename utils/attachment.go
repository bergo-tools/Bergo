package utils

import (
	"bergo/locales"
	"bytes"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
)

const (
	AttachmentTypeFile = iota
	AttachmentTypeDir
	AttachmentTypeImage
)

// 支持的图片格式及其 MIME 类型
var imageExtensions = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".gif":  "image/gif",
	".webp": "image/webp",
}

type Attachment struct {
	Index int    `json:"index"`
	Path  string `json:"path"`
	Type  int    `json:"type"`
}

// IsImageFile 检查文件是否为支持的图片格式
func IsImageFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	_, ok := imageExtensions[ext]
	return ok
}

// GetImageMimeType 获取图片的 MIME 类型
func GetImageMimeType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	if mime, ok := imageExtensions[ext]; ok {
		return mime
	}
	return "image/jpeg" // 默认
}

// GetImageDataURL 读取图片并返回 base64 data URL
func (a *Attachment) GetImageDataURL() (string, error) {
	return GetImageDataURL(a.Path)
}

// GetImageDataURL 根据路径读取图片并返回 base64 data URL
func GetImageDataURL(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	mime := GetImageMimeType(path)
	base64Data := base64.StdEncoding.EncodeToString(data)
	return "data:" + mime + ";base64," + base64Data, nil
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
	case AttachmentTypeImage:
		// 图片附件只返回描述，实际图片数据通过 GetImageDataURL 获取
		buf.WriteString(locales.Sprintf("%v. image %v submitted as attachment\n", a.Index, a.Path))
	}
	return buf.String()
}
