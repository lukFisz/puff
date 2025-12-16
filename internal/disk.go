package internal

import (
	"fmt"
	"syscall"
)

func DiskFreeSpaceInBytes(path string) (*uint64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return nil, fmt.Errorf("error reading stats of %s, error: %s", path, err)
	}
	free := stat.Bavail * uint64(stat.Bsize)
	return &free, nil
}
