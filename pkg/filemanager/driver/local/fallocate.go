//go:build !linux && !darwin
// +build !linux,!darwin

package local

import "os"

// No-op on non-Linux/Darwin platforms.
func Fallocate(file *os.File, offset int64, length int64) error {
	return nil
}
