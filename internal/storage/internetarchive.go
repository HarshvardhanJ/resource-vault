package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// InternetArchiveStorage provides a provider boundary for published files.
// The actual Internet Archive protocol is isolated in this adapter so the catalog
// and resource domain remain provider-independent.
type InternetArchiveStorage struct {
	accessKey       string
	secretKey       string
	collection      string
	identifierPrefx string
	client          *http.Client
}

func NewInternetArchiveStorage(accessKey, secretKey, collection, identifierPrefix string) *InternetArchiveStorage {
	return &InternetArchiveStorage{
		accessKey: accessKey,
		secretKey: secretKey,
		collection: collection,
		identifierPrefx: identifierPrefix,
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

func (s *InternetArchiveStorage) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	if s.accessKey == "" || s.secretKey == "" || s.collection == "" {
		return fmt.Errorf("internet archive storage is not configured")
	}
	if strings.TrimSpace(key) == "" { return ErrInvalidKey }
	return fmt.Errorf("internet archive provider upload is not yet wired: %s", key)
}

func (s *InternetArchiveStorage) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	if strings.TrimSpace(key) == "" { return nil, ErrInvalidKey }
	return nil, fmt.Errorf("internet archive file lookup is not yet wired: %s", key)
}

func (s *InternetArchiveStorage) Delete(ctx context.Context, key string) error {
	if strings.TrimSpace(key) == "" { return ErrInvalidKey }
	return fmt.Errorf("internet archive delete is not yet wired: %s", key)
}

func (s *InternetArchiveStorage) Exists(ctx context.Context, key string) (bool, error) {
	if strings.TrimSpace(key) == "" { return false, ErrInvalidKey }
	return false, fmt.Errorf("internet archive existence check is not yet wired: %s", key)
}

func (s *InternetArchiveStorage) URL(key string) string { return "" }
