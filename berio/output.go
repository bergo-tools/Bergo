package berio

const (
	MsgTypeText = iota
	MsgTypeWarning
	MsgTypeProgressBar
	MsgTypeTodoList
	MsgTypeDump
)

type ProgressBar struct {
	Total int
	Incr  int
	Title string
}

// 用于监听流程的输出信息
type BerOutput interface {
	OnLLMResponse(response string, isReasoning bool)
	Stop() string
	OnSystemMsg(msg interface{}, typ int)
	UpdateTail(tail string)
}
