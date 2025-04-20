package local

import (
	"os"
	"syscall"
	"unsafe"
)

func Fallocate(file *os.File, offset int64, length int64) error {
	var fst syscall.Fstore_t

	fst.Flags = syscall.F_ALLOCATECONTIG
	fst.Posmode = syscall.F_PREALLOCATE
	fst.Offset = 0
	fst.Length = offset + length
	fst.Bytesalloc = 0

	// Check https://lists.apple.com/archives/darwin-dev/2007/Dec/msg00040.html
	_, _, err := syscall.Syscall(syscall.SYS_FCNTL, file.Fd(), syscall.F_PREALLOCATE, uintptr(unsafe.Pointer(&fst)))
	if err != syscall.Errno(0x0) {
		fst.Flags = syscall.F_ALLOCATEALL
		// Ignore the return value
		_, _, _ = syscall.Syscall(syscall.SYS_FCNTL, file.Fd(), syscall.F_PREALLOCATE, uintptr(unsafe.Pointer(&fst)))
	}

	return syscall.Ftruncate(int(file.Fd()), fst.Length)
}
