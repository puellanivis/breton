// +build dragonflybsd freebsd linux netbsd openbsd solaris darwin

package files

import (
	"os"
	"strconv"
)

func resolveFileHandle(num string) (uintptr, error) {
	fd, err := strconv.ParseUint(num, 0, strconv.IntSize)
	if err != nil {
		return uintptr(^fd), os.ErrInvalid
	}

	return uintptr(fd), nil
}
