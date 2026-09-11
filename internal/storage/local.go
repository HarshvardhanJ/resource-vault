package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type LocalStorage struct {
	baseDir string
	baseURL string
}

func NewLocalStorage(baseDir string, baseURL string) (*LocalStorage, error) {
	cleanDir := filepath.Clean(baseDir)
	if err := os.MkdirAll(cleanDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create local storage directory: %w", err)
	}
	return &LocalStorage{
		baseDir: cleanDir,
		baseURL: strings.TrimRight(baseURL, "/"),
	}, nil
}

func (s *LocalStorage) resolvePath(key string) (string, error) {
	if strings.Contains(key, "..") {
		return "", ErrInvalidKey
	}
	cleanKey := filepath.Clean(key)
	cleanKey = strings.TrimPrefix(cleanKey, string(filepath.Separator))

	targetPath := filepath.Join(s.baseDir, cleanKey)
	rel, err := filepath.Rel(s.baseDir, targetPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", ErrInvalidKey
	}
	return targetPath, nil
}

func (s *LocalStorage) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	targetPath, err := s.resolvePath(key)
	if err != nil {
		return err
	}

	dir := filepath.Dir(targetPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	tmpFile, err := os.CreateTemp(dir, "upload-*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpName := tmpFile.Name()
	defer os.Remove(tmpName)

	written, err := io.Copy(tmpFile, r)
	if err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write object bytes: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	if size > 0 && written != size {
		return fmt.Errorf("byte count mismatch: expected %d, got %d", size, written)
	}

	if err := os.Rename(tmpName, targetPath); err != nil {
		return fmt.Errorf("failed to commit file to storage: %w", err)
	}

	return nil
}

func (s *LocalStorage) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	targetPath, err := s.resolvePath(key)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return f, nil
}

func (s *LocalStorage) Delete(ctx context.Context, key string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	targetPath, err := s.resolvePath(key)
	if err != nil {
		return err
	}

	if err := os.Remove(targetPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (s *LocalStorage) Exists(ctx context.Context, key string) (bool, error) {
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}

	targetPath, err := s.resolvePath(key)
	if err != nil {
		return false, err
	}

	_, err = os.Stat(targetPath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (s *LocalStorage) URL(key string) string {
	if s.baseURL == "" {
		return ""
	}
	return fmt.Sprintf("%s/storage/%s", s.baseURL, strings.TrimPrefix(key, "/"))
}
