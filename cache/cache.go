package cache

import (
	"context"
	"errors"
	"io"
)

var ErrNotFound = errors.New("cache: not found")

// KeyFunc is a function that writes the key to the provided writer for hashing.
// It is to prevent the need to allocate a []byte slice for the key.
//
// KeyFunc must be deterministic and must write identical bytes each time it is called.
type KeyFunc func(w io.Writer) error

// Cache is an interface for a simple key-value store for caching data.
type Cache interface {
	// Open returns a reader for the given key. If the key does not exist, it returns [ErrNotFound].
	Open(ctx context.Context, key KeyFunc) (io.ReadCloser, error)

	// Create returns a writer for the given key. If the key already exists, it should be overwritten.
	Create(ctx context.Context, key KeyFunc) (BlobWriter, error)
}

// BlobWriter is an interface for writing data to the cache. It is returned by [Cache.Create] and should be closed when done.
type BlobWriter interface {
	io.WriteCloser

	// Abort aborts the write operation and discards any temporary data.
	Abort() error
}
