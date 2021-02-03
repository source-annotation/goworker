package goworker

// job 的消息体
// class : 消息名
// args : 参数
type Payload struct {
	Class string        `json:"class"`
	Args  []interface{} `json:"args"`
}
