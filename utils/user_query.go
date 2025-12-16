package utils

import (
	"bergo/prompt"

	"bytes"
)

type Query struct {
	UserInput     string        `json:"user_input"`
	Mode          string        `json:"mode"`
	Attachment    []*Attachment `json:"attachment"`
	IsInterrupted bool          `json:"is_interrupted"`
	MememtoNotice bool          `json:"mememto_notice"`
	IsCompact     bool
}

func (q *Query) SetCompact() {
	q.IsCompact = true
}
func (q *Query) SetMememtoNotice() {
	q.MememtoNotice = true
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

	if q.MememtoNotice {
		buf.WriteString("<mememto_notice>记得维护mememto file，在开始任务时先写入，之后不断更新它，当然，在Debug模式下不需要维护</mememto_notice>\n")
		q.MememtoNotice = false
	}
	return buf.String()
}
