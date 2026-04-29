//go:build !windows

package atomicfile

import "os"

func atomicRename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}
