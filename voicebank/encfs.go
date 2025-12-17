package voicebank

/*

import (
	"fmt"
	"io/fs"

	"golang.org/x/text/encoding"
)


type encFS struct {
	fs.FS
	enc encoding.Encoding
}

func (f *encFS) Open(name string) (fs.File, error) {
	encName, err := f.enc.NewEncoder().String(name)
	if err != nil {
		return nil, fmt.Errorf("failed to encode name: %w", err)
	}

	file, err := f.FS.Open(encName)
	if err != nil {
		return nil, err
	}

	return &encFSFile{File: file, enc: f.enc}, nil
}

func (f *encFS) Glob(pattern string) ([]string, error) {
	encPattern, err := f.enc.NewEncoder().String(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to encode name: %w", err)
	}

	matches, err := fs.Glob(f.FS, encPattern)
	if err != nil {
		return nil, err
	}
	dec := f.enc.NewDecoder()
	for i := range matches {
		decMatch, err := dec.String(matches[i])
		if err != nil {
			return nil, fmt.Errorf("failed to decode match: %w", err)
		}
		matches[i] = decMatch
	}
	return matches, nil
}

func (f *encFS) ReadDir(name string) ([]fs.DirEntry, error) {
	encName, err := f.enc.NewEncoder().String(name)
	if err != nil {
		return nil, fmt.Errorf("failed to encode name: %w", err)
	}

	entries, err := fs.ReadDir(f.FS, encName)
	if err != nil {
		return nil, err
	}

	out := make([]fs.DirEntry, len(entries))
	for i, e := range entries {
		out[i] = &encDirEntry{DirEntry: e, enc: f.enc}
	}

	return out, nil
}

type encFSFile struct {
	fs.File
	enc encoding.Encoding
}

func (f *encFSFile) Stat() (fs.FileInfo, error) {
	fi, err := f.File.Stat()
	if err != nil {
		return nil, err
	}
	return &encFSFileInfo{FileInfo: fi, enc: f.enc}, nil
}

func (f *encFSFile) ReadDir(n int) ([]fs.DirEntry, error) {
	rdf, ok := f.File.(fs.ReadDirFile)
	if !ok {
		return nil, fs.ErrInvalid
	}

	entries, err := rdf.ReadDir(n)
	if err != nil {
		return nil, err
	}

	out := make([]fs.DirEntry, len(entries))
	for i, e := range entries {
		out[i] = &encDirEntry{DirEntry: e, enc: f.enc}
	}

	return out, nil
}

type encDirEntry struct {
	fs.DirEntry
	enc encoding.Encoding
}

func (e *encDirEntry) Name() string {
	name, _ := e.enc.NewDecoder().String(e.DirEntry.Name())
	return name
}

type encFSFileInfo struct {
	fs.FileInfo
	enc encoding.Encoding
}

func (fi *encFSFileInfo) Name() string {
	name, _ := fi.enc.NewDecoder().String(fi.FileInfo.Name())
	return name
}
*/
