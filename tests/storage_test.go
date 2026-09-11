package tests

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/nitc-pyq-archive/archive/internal/storage"
)

func TestLocalStorage(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "storage-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store, err := storage.NewLocalStorage(tempDir, "http://localhost:8080")
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	ctx := context.Background()
	content := []byte("%PDF-1.4 test document content")
	key := "cs201/2025/endsem.pdf"

	// 1. Put
	err = store.Put(ctx, key, bytes.NewReader(content), int64(len(content)), "application/pdf")
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// 2. Exists
	exists, err := store.Exists(ctx, key)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Errorf("expected key to exist, but Exists returned false")
	}

	// 3. Get
	r, err := store.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer r.Close()

	readBytes, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}
	if !bytes.Equal(readBytes, content) {
		t.Errorf("expected %q, got %q", content, readBytes)
	}

	// 4. URL
	url := store.URL(key)
	expectedURL := "http://localhost:8080/storage/cs201/2025/endsem.pdf"
	if url != expectedURL {
		t.Errorf("expected URL %q, got %q", expectedURL, url)
	}

	// 5. Path traversal guard
	err = store.Put(ctx, "../../evil.pdf", bytes.NewReader(content), int64(len(content)), "application/pdf")
	if err == nil {
		t.Errorf("expected error on path traversal key, but got nil")
	}

	// 6. Delete
	err = store.Delete(ctx, key)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	exists, err = store.Exists(ctx, key)
	if err != nil {
		t.Fatalf("Exists after delete failed: %v", err)
	}
	if exists {
		t.Errorf("expected key to be deleted, but still exists")
	}
}
