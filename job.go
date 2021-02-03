package goworker

// job
type Job struct {
	Queue   string  // job 所属的  queue
	Payload Payload // 该 job 的消息体
}
