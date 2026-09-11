package tests

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/nitc-pyq-archive/archive/internal/resources"
)

func TestAcademicYearDisplay(t *testing.T) {
	tests := []struct {
		year     int
		expected string
	}{
		{2025, "2025-26"},
		{2024, "2024-25"},
		{2029, "2029-30"},
		{1999, "1999-00"},
		{0, ""},
		{-1, ""},
	}

	for _, tc := range tests {
		res := resources.Resource{AcademicYearStart: tc.year}
		actual := res.AcademicYearDisplay()
		if actual != tc.expected {
			t.Errorf("for year %d: expected %q, got %q", tc.year, tc.expected, actual)
		}
	}
}

func TestFileSizeDisplay(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1048576, "1.0 MB"},
		{5242880, "5.0 MB"},
	}

	for _, tc := range tests {
		res := resources.Resource{FileSizeBytes: tc.bytes}
		actual := res.FileSizeDisplay()
		if actual != tc.expected {
			t.Errorf("for %d bytes: expected %q, got %q", tc.bytes, tc.expected, actual)
		}
	}
}

func TestPDFValidationLogic(t *testing.T) {
	validPDF := []byte("%PDF-1.4\n1 0 obj\n...")
	invalidPDF := []byte("<!DOCTYPE html><html>...")
	emptyFile := []byte("")

	isPDF := func(data []byte) bool {
		return len(data) >= 5 && bytes.HasPrefix(data, []byte("%PDF-"))
	}

	if !isPDF(validPDF) {
		t.Errorf("expected validPDF to pass validation")
	}
	if isPDF(invalidPDF) {
		t.Errorf("expected invalidPDF to fail validation")
	}
	if isPDF(emptyFile) {
		t.Errorf("expected emptyFile to fail validation")
	}
}

func TestSHA256Computation(t *testing.T) {
	data := []byte("nitc pyq test content")
	hash := sha256.Sum256(data)
	hashStr := hex.EncodeToString(hash[:])

	if len(hashStr) != 64 {
		t.Fatalf("expected 64 character hex string, got %d", len(hashStr))
	}

	// Verify idempotency
	hash2 := sha256.Sum256(data)
	if hashStr != hex.EncodeToString(hash2[:]) {
		t.Errorf("SHA-256 computation not deterministic")
	}
}
