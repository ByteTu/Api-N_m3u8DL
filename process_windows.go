//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

const (
	PROCESS_SUSPEND_RESUME            = 0x0800
	PROCESS_TERMINATE                 = 0x0001
	PROCESS_QUERY_LIMITED_INFORMATION = 0x1000
)

var (
	modkernel32          = syscall.NewLazyDLL("kernel32.dll")
	procNtSuspendProcess = modkernel32.NewProc("NtSuspendProcess")
	procNtResumeProcess  = modkernel32.NewProc("NtResumeProcess")
	procTerminateProcess = modkernel32.NewProc("TerminateProcess")
)

func suspendProcess(pid int) error {
	handle, err := syscall.OpenProcess(PROCESS_SUSPEND_RESUME, false, uint32(pid))
	if err != nil {
		return err
	}
	defer syscall.CloseHandle(handle)
	_, _, err = procNtSuspendProcess.Call(uintptr(handle))
	return err
}

func resumeProcess(pid int) error {
	handle, err := syscall.OpenProcess(PROCESS_SUSPEND_RESUME, false, uint32(pid))
	if err != nil {
		return err
	}
	defer syscall.CloseHandle(handle)
	_, _, err = procNtResumeProcess.Call(uintptr(handle))
	return err
}

func terminateProcess(pid int, exitCode int) error {
	handle, err := syscall.OpenProcess(PROCESS_TERMINATE, false, uint32(pid))
	if err != nil {
		return err
	}
	defer syscall.CloseHandle(handle)
	_, _, err = procTerminateProcess.Call(uintptr(handle), uintptr(exitCode))
	return err
}

func isProcessRunning(pid int) bool {
	handle, err := syscall.OpenProcess(PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return false
	}
	defer syscall.CloseHandle(handle)

	var exitCode uint32
	ret, _, _ := syscall.NewLazyDLL("kernel32.dll").NewProc("GetExitCodeProcess").Call(uintptr(handle), uintptr(unsafe.Pointer(&exitCode)))
	if ret == 0 {
		return false
	}
	return exitCode == 259
}
