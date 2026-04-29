// Package memcache implements a in-memory cache.
// It uses xxHash3 64-bit hashing to store and retrieve blobs of data. It is not persistent.
package memcache

import (
	"bytes"
	"context"
	"io"
	"sync"

	"github.com/SladkyCitron/gotau/cache"
	"github.com/zeebo/xxh3"
)

var _ cache.Cache = (*Cache)(nil)

type Cache struct {
	blobs map[uint64][]byte
	mu    sync.RWMutex
}

func New() *Cache {
	return &Cache{blobs: make(map[uint64][]byte)}
}

func (c *Cache) hash(key cache.KeyFunc) uint64 {
	h := xxh3.New()
	key(h)
	return h.Sum64()
}

func (c *Cache) Open(ctx context.Context, key cache.KeyFunc) (io.ReadCloser, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	hash := c.hash(key)

	c.mu.RLock()
	blob, ok := c.blobs[hash]
	newBlob := make([]byte, len(blob))
	copy(newBlob, blob)
	c.mu.RUnlock()
	if !ok {
		return nil, cache.ErrNotFound
	}

	return io.NopCloser(bytes.NewReader(newBlob)), nil
}

func (c *Cache) Create(ctx context.Context, key cache.KeyFunc) (cache.BlobWriter, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	hash := c.hash(key)

	return &blobWriter{c: c, hash: hash}, nil
}

type blobWriter struct {
	c    *Cache
	hash uint64
	buf  Buffer
}

func (w *blobWriter) Write(p []byte) (int, error) {
	return w.buf.Write(p)
}

func (w *blobWriter) Seek(offset int64, whence int) (int64, error) {
	return w.buf.Seek(offset, whence)
}

func (w *blobWriter) Close() error {
	w.c.mu.Lock()
	defer w.c.mu.Unlock()
	w.c.blobs[w.hash] = w.buf.Bytes()
	return nil
}

func (w *blobWriter) Abort() error {
	w.buf.Reset()
	return nil
}
