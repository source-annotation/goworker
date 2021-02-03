// +build !go1.1

// 条件编译，go 1.1 以下会编译该文件
package goworker

import (
	"os"
)

// Stops signals channel. This does not exist in
// Go less than 1.1.
func signalStop(c chan<- os.Signal) {
}
