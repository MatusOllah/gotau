package cache

import (
	"context"
	"io"
)

var _ Cache = (*NopCache)(nil)

// NopCache is a cache that doesn't do anything. It always returns [ErrNotFound] for [Open] and does nothing for [Create].
type NopCache struct{}

func (c *NopCache) Open(_ context.Context, _ KeyFunc) (io.ReadCloser, error) {
	return nil, ErrNotFound
}

func (c *NopCache) Create(_ context.Context, _ KeyFunc) (BlobWriter, error) {
	return &nopBlobWriter{}, nil
}

type nopBlobWriter struct{}

func (c *nopBlobWriter) Write(_ []byte) (int, error) {
	return 0, nil
}

func (c *nopBlobWriter) Seek(_ int64, _ int) (int64, error) {
	return 0, nil
}

func (c *nopBlobWriter) Close() error {
	return nil
}

func (c *nopBlobWriter) Abort() error {
	return nil
}
