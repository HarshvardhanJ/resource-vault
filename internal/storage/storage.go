package storage

import (
	"context"
	"errors"
	"io"
)

var (
	ErrNotFound = errors.New("object not found")
	ErrInvalidKey = errors.New("invalid storage key")
)

// ObjectStore provides an abstraction over published file byte storage (local disk, Internet Archive, S3).
type ObjectStore interface {
	// Put stores an object with key, content reader, content size, and content type.
	Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error

	// Get retrieves an object reader. Caller is responsible for closing the reader.
	Get(ctx context.Context, key string) (io.ReadCloser, error)

	// Delete removes an object by key.
	Delete(ctx context.Context, key string) error

	// Exists checks if an object exists.
	Exists(ctx context.Context, key string) (bool, error)

	// URL returns a public or resolvable URL for the given key, or empty string if streaming is required.
	URL(key string) string
}
