//go:build windows

package atomicfile

import (
	"os"

	"golang.org/x/sys/windows"
)

func atomicRename(oldpath, newpath string) error {
	src, err := windows.UTF16PtrFromString(oldpath)
	if err != nil {
		return &os.LinkError{Op: "rename", Old: oldpath, New: newpath, Err: err}
	}
	dst, err := windows.UTF16PtrFromString(newpath)
	if err != nil {
		return &os.LinkError{Op: "rename", Old: oldpath, New: newpath, Err: err}
	}
	if err := windows.MoveFileEx(src, dst, windows.MOVEFILE_REPLACE_EXISTING|windows.MOVEFILE_WRITE_THROUGH); err != nil {
		return &os.LinkError{Op: "rename", Old: oldpath, New: newpath, Err: err}
	}
	return nil
}
