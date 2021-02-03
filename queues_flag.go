package goworker

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	errorEmptyQueues      = errors.New("you must specify at least one queue")
	errorNonNumericWeight = errors.New("the weight must be a numeric value")
)

type queuesFlag []string

// 通过命令行参数传入 queue ，支持设置权重。
// 例如 -queue="a=2,b=1"， queue a 就会被 append 两个次到 workerSetting.Queues，b 只有1次。
// 此时 workerSettings.Queues = [a, a, b]， poller 从 queue a 取 job 的概率就是 queue b 的2倍。
func (q *queuesFlag) Set(value string) error {
	for _, queueAndWeight := range strings.Split(value, ",") {
		if queueAndWeight == "" {
			continue
		}

		queue, weight, err := parseQueueAndWeight(queueAndWeight)
		if err != nil {
			return err
		}

		for i := 0; i < weight; i++ {
			*q = append(*q, queue)
		}
	}
	if len(*q) == 0 {
		return errorEmptyQueues
	}
	return nil
}

func (q *queuesFlag) String() string {
	return fmt.Sprint(*q)
}

func parseQueueAndWeight(queueAndWeight string) (queue string, weight int, err error) {
	parts := strings.SplitN(queueAndWeight, "=", 2)
	queue = parts[0]

	if queue == "" {
		return
	}

	if len(parts) == 1 {
		weight = 1
	} else {
		weight, err = strconv.Atoi(parts[1])
		if err != nil {
			err = errorNonNumericWeight
		}
	}
	return
}
