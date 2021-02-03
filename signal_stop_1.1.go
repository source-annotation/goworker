// +build go1.1

// 条件编译，go 1.1 及以上会编译该文件
package goworker

import (
	"os"
	"os/signal"
)

// Stops signals channel. This function exists
// in Go greater or equal to 1.1.
func signalStop(c chan<- os.Signal) {
	signal.Stop(c)
}
