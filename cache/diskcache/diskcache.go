// Package diskcache implements an on-disk cache.
// It uses xxHash3 128-bit hashing and a content-addressable storage system with
// 1-level directory sharding to store and retrieve blobs of data. It is persistent.
package diskcache

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/SladkyCitron/gotau/cache"
	"github.com/SladkyCitron/gotau/cache/diskcache/internal/atomicfile"
	"github.com/zeebo/xxh3"
)

var _ cache.Cache = (*Cache)(nil)

type Cache struct {
	basePath string
	ext      string
}

func New(path string, ext string) *Cache {
	return &Cache{basePath: path, ext: ext}
}

func (c *Cache) hash(key cache.KeyFunc) xxh3.Uint128 {
	h := xxh3.New128()
	key(h)
	return h.Sum128()
}

func (c *Cache) pathForHash(hash xxh3.Uint128) string {
	hashBytes := hash.Bytes()
	hashString := hex.EncodeToString(hashBytes[:])
	return filepath.Join(c.basePath, hashString[:2], hashString+c.ext)
}

func (c *Cache) Open(ctx context.Context, key cache.KeyFunc) (io.ReadCloser, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	hash := c.hash(key)
	path := c.pathForHash(hash)

	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, cache.ErrNotFound
		} else {
			return nil, fmt.Errorf("diskcache: failed to open cached file: %w", err)
		}
	}

	select {
	case <-ctx.Done():
		_ = f.Close()
		return nil, ctx.Err()
	default:
		return f, nil
	}
}

func (c *Cache) Create(ctx context.Context, key cache.KeyFunc) (cache.BlobWriter, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	hash := c.hash(key)
	path := c.pathForHash(hash)

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("diskcache: failed to create cache directory: %w", err)
	}

	f, err := atomicfile.Create(path, 0666)
	if err != nil {
		return nil, fmt.Errorf("diskcache: failed to create file: %w", err)
	}

	select {
	case <-ctx.Done():
		_ = f.Abort()
		return nil, ctx.Err()
	default:
		return f, nil
	}
}

// Dir returns the default directory for the disk cache based on the user's cache directory and the subdirectory.
func Dir(name string) (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheDir, name), nil
}
