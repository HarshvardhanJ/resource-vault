package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// InternetArchiveStorage is a deliberately small provider adapter. The exact IA
// API/credential flow is kept behind this type so the catalog does not know about
// Internet Archive details.
//
// The production implementation should be completed using the current Internet
// Archive S3-like upload/download API. Until credentials are configured, the app
// should fail closed rather than silently falling back to local disk in production.
type InternetArchiveStorage struct {
	accessKey       string
	secretKey       string
	collection      string
	identifierPrefx string
	client          *http.Client
}

func NewInternetArchiveStorage(accessKey, secretKey, collection, identifierPrefix string) *InternetArchiveStorage {
	return &InternetArchiveStorage{
		accessKey:       accessKey,
		secretKey:       secretKey,
		collection:      collection,
		identifierPrefx: identifierPrefix,
		client:          &http.Client{Timeout: 60 * time.Second},
	}
}

func (s *InternetArchiveStorage) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	if s.accessKey == "" || s.secretKey == "" || s.collection == "" {
		return fmt.Errorf("internet archive storage is not configured")
	}
	if strings.TrimSpace(key) == "" {
		return ErrInvalidKey
	}
	// Provider-specific upload is intentionally isolated here. The implementation
	// must use server-side credentials only and must never expose them to browsers.
	return fmt.Errorf("internet archive upload adapter not yet implemented for key %s", key)
}

func (s *InternetArchiveStorage) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	if strings.TrimSpace(key) == "" {
		return nil, ErrInvalidKey
	}
	return nil, fmt.Errorf("internet archive Get requires configured item/file mapping for %s", key)
}

func (s *InternetArchiveStorage) Delete(ctx context.Context, key string) error {
	if strings.TrimSpace(key) == "" {
		return ErrInvalidKey
	}
	return fmt.Errorf("internet archive Delete is intentionally explicit and not implemented for %s", key)
}

func (s *InternetArchiveStorage) Exists(ctx context.Context, key string) (bool, error) {
	if strings.TrimSpace(key) == "" {
		return false, ErrInvalidKey
	}
	return false, fmt.Errorf("internet archive Exists requires configured item/file mapping for %s", key)
}

func (s *InternetArchiveStorage) URL(key string) string {
	if s.collection == "" || strings.TrimSpace(key) == "" {
		return ""
	}
	// This helper is only a resolver placeholder. URL construction must ultimately
	// come from the canonical IA identifier and filename stored in PostgreSQL.
	return ""
}

// iaQueryEscape is retained for the eventual provider implementation.
func iaQueryEscape(v string) string { return url.QueryEscape(v) }
