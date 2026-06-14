//go:build windows

package tree

import "errors"

func createFIFO(path string) error {
	return errors.New("FIFO is not supported on Windows")
}
