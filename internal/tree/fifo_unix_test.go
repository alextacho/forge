//go:build !windows

package tree

import "syscall"

func createFIFO(path string) error {
	return syscall.Mkfifo(path, 0o600)
}
