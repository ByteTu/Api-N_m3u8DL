//go:build linux || darwin

package main

import (
	"os"
	"syscall"
)

func suspendProcess(pid int) error {
	return syscall.Kill(pid, syscall.SIGSTOP)
}

func resumeProcess(pid int) error {
	return syscall.Kill(pid, syscall.SIGCONT)
}

func terminateProcess(pid int, exitCode int) error {
	return syscall.Kill(pid, syscall.SIGKILL)
}

func isProcessRunning(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// 发送信号0来检查进程是否存在（不会实际发送信号）
	err = proc.Signal(syscall.Signal(0))
	return err == nil
}
