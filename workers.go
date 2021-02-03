package goworker

import (
	"encoding/json"
	"fmt"
)

var (
	// map[class]func
	// 例如 map["work_SendEmail"]FuncSendEmail
	workers map[string]workerFunc
)

func init() {
	workers = make(map[string]workerFunc)
}

// Register registers a goworker worker function. Class
// refers to the Ruby name of the class which enqueues the
// job. Worker is a function which accepts a queue and an
// arbitrary array of interfaces as arguments.
func Register(class string, worker workerFunc) {
	workers[class] = worker
}

func Enqueue(job *Job) error {
	err := Init()
	if err != nil {
		return err
	}

	conn, err := GetConn()
	if err != nil {
		logger.Criticalf("Error on getting connection on enqueue")
		return err
	}
	defer PutConn(conn)

	buffer, err := json.Marshal(job.Payload)
	if err != nil {
		logger.Criticalf("Cant marshal payload on enqueue")
		return err
	}

	err = conn.Send("RPUSH", fmt.Sprintf("%squeue:%s", workerSettings.Namespace, job.Queue), buffer)
	if err != nil {
		logger.Criticalf("Cant push to queue")
		return err
	}

	// 把 queue 加入到 resque:squques(SET) 中，如果已存在则忽略(集合元素不重复)
	err = conn.Send("SADD", fmt.Sprintf("%squeues", workerSettings.Namespace), job.Queue)
	if err != nil {
		logger.Criticalf("Cant register queue to list of use queues")
		return err
	}

	return conn.Flush()
}
