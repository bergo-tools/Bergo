package utils

import (
	"bergo/prompt"

	"bytes"
)

type Query struct {
	UserInput           string        `json:"user_input"`
	Mode                string        `json:"mode"`
	Attachment          []*Attachment `json:"attachment"`
	IsInterrupted       bool          `json:"is_interrupted"`
	MementoUpdateRemind bool          `json:"memento_update_remind"`
	IsCompact           bool
}

func (q *Query) SetCompact() {
	q.IsCompact = true
}
func (q *Query) SetMementoUpdateRemind() {
	q.MementoUpdateRemind = true
}

func (q *Query) SetAttachment(attachments []*Attachment) {
	q.Attachment = append(q.Attachment, attachments...)
}

func (q *Query) SetUserInput(input string) {
	q.UserInput = input
}

func (q *Query) GetUserInput() string {
	return q.UserInput
}
func (q *Query) SetMode(mode string) {
	q.Mode = mode
}

func (q *Query) Interrupt() {
	q.IsInterrupted = true
}

// GetImageDataURL 获取第一个图片附件的 data URL（目前只支持单张图片）
func (q *Query) GetImageDataURL() string {
	for _, attachment := range q.Attachment {
		if attachment.Type == AttachmentTypeImage {
			dataURL, err := attachment.GetImageDataURL()
			if err == nil {
				return dataURL
			}
		}
	}
	return ""
}

func (q *Query) Build() string {
	buf := bytes.NewBuffer(nil)

	//USER INPUT
	if q.IsInterrupted {
		buf.WriteString("用户终止了行动，或者发生了致命错误，请按照下面的用户指令重新开始任务\n")
	}
	if q.IsCompact {
		buf.WriteString("**用户进行了上下文压缩，你之前的上下文已经没有了。请查看memento file了解前情再进行用户下面的指令**\n")
	}
	buf.WriteString(prompt.GetModePrompt(q.Mode))
	buf.WriteString("\n")
	buf.WriteString(NewTagContent(q.UserInput, "user_input").WholeContent)
	buf.WriteString("\n")
	if len(q.Attachment) > 0 {
		buf.WriteString("<attachments>\n")
		for _, attachment := range q.Attachment {
			buf.WriteString(attachment.GetContent())
		}
		buf.WriteString("</attachments>\n")
	}
	if q.MementoUpdateRemind {
		buf.WriteString("<memento_update_remind>检测到你在上一轮任务中没有更新memento file，请记得及时更新它以保存任务进度和关键信息</memento_update_remind>\n")
		q.MementoUpdateRemind = false
	}
	return buf.String()
}
