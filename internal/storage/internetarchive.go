package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// InternetArchiveStorage is the provider boundary for published archive files.
// Credentials are server-side only. The actual IA protocol is deliberately isolated
// so catalog/resource code remains provider-independent.
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
	if s.accessKey == "" || s.secretKey == "" || s.collection == "" { return fmt.Errorf("internet archive storage is not configured") }
	if strings.TrimSpace(key) == "" { return ErrInvalidKey }
	return fmt.Errorf("internet archive provider upload is not yet wired: %s", key)
}
func (s *InternetArchiveStorage) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	if strings.TrimSpace(key) == "" { return nil, ErrInvalidKey }
	return nil, fmt.Errorf("internet archive file lookup not yet wired: %s", key)
}
func (s *InternetArchiveStorage) Delete(ctx context.Context, key string) error {
	if strings.TrimSpace(key) == "" { return ErrInvalidKey }
	return fmt.Errorf("internet archive delete not yet wired: %s", key)
}
func (s *InternetArchiveStorage) Exists(ctx context.Context, key string) (bool, error) {
	if strings.TrimSpace(key) == "" { return false, ErrInvalidKey }
	return false, fmt.Errorf("internet archive existence check not yet wired: %s", key)
}
func (s *InternetArchiveStorage) URL(key string) string { return "" }
