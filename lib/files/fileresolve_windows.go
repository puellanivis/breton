package files

import (
	"os"
	"strconv"
)

func resolveFileHandle(num string) (uintptr, error) {
	fd, err := strconv.ParseInt(num, 0, 32)
	if err != nil {
		return uintptr(^fd), os.ErrInvalid
	}

	return uintptr(fd), nil
}
