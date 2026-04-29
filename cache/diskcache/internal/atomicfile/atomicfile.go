// Package atomicfile implements atomic files.
package atomicfile

import (
	"fmt"
	"os"
	"path/filepath"
)

// File is like os.File but instead it creates a temporary file for writing and then does
// an atomic rename on [File.Close].
type File struct {
	*os.File
	path string
	dir  string
	done bool
}

// Create creates a new [File].
func Create(path string, mode os.FileMode) (*File, error) {
	dir, file := filepath.Split(path)
	f, err := os.CreateTemp(dir, "."+file+".tmp-*")
	if err != nil {
		return nil, err
	}
	if err := f.Chmod(mode); err != nil {
		// clean up
		_ = f.Close()
		_ = os.Remove(f.Name())
		return nil, err
	}
	return &File{File: f, path: path, dir: dir}, nil
}

// Close commits the file by atomic renaming the temporary file to the actual path.
func (f *File) Close() error {
	if f.done {
		return nil
	}
	if err := f.File.Sync(); err != nil {
		_ = f.File.Close()
		_ = os.Remove(f.Name())
		return err
	}
	if err := f.File.Close(); err != nil {
		_ = os.Remove(f.Name())
		return err
	}
	if err := atomicRename(f.Name(), f.path); err != nil {
		_ = os.Remove(f.Name())
		return err
	}
	if err := syncDir(f.dir); err != nil {
		return fmt.Errorf("failed to sync directory: %w", err)
	}
	f.done = true
	return nil
}

// Abort aborts the write and removes the temporary file.
func (f *File) Abort() error {
	if f.done {
		return nil
	}
	if err := f.File.Close(); err != nil {
		_ = os.Remove(f.Name())
		return err
	}
	if err := os.Remove(f.Name()); err != nil {
		return err
	}
	f.done = true
	return nil
}

func syncDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	return d.Sync()
}
