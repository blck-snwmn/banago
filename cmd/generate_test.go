package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/genai"
)

func TestNormalizeExt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		mimeType string
		want     string
	}{
		{"jpeg", "image/jpeg", ".jpg"},
		{"jpg alias", "image/jpg", ".jpg"},
		{"png", "image/png", ".png"},
		{"webp", "image/webp", ".webp"},
		{"gif", "image/gif", ".gif"},
		{"bmp", "image/bmp", ".bmp"},
		{"avif", "image/avif", ".avif"},
		{"heic", "image/heic", ".heic"},
		{"heif", "image/heif", ".heif"},
		{"tiff", "image/tiff", ".tiff"},
		{"tif alias", "image/tif", ".tiff"},
		{"jpeg uppercase", "IMAGE/JPEG", ".jpg"},
		{"jpeg with params", "image/jpeg; charset=utf-8", ".jpg"},
		{"unknown", "application/octet-stream", ".bin"},
		{"empty", "", ".bin"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := normalizeExt(tt.mimeType)
			if got != tt.want {
				t.Errorf("normalizeExt(%q) = %q, want %q", tt.mimeType, got, tt.want)
			}
		})
	}
}

func TestImagePartFromFile(t *testing.T) {
	t.Parallel()

	t.Run("valid png image", func(t *testing.T) {
		t.Parallel()
		// 1x1 transparent PNG
		pngData := []byte{
			0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
			0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
			0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
			0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4,
			0x89, 0x00, 0x00, 0x00, 0x0A, 0x49, 0x44, 0x41,
			0x54, 0x78, 0x9C, 0x63, 0x00, 0x01, 0x00, 0x00,
			0x05, 0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00,
			0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE,
			0x42, 0x60, 0x82,
		}

		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "test.png")
		if err := os.WriteFile(tmpFile, pngData, 0o644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		part, err := imagePartFromFile(tmpFile)
		if err != nil {
			t.Fatalf("imagePartFromFile() error = %v", err)
		}
		if part == nil || part.InlineData == nil {
			t.Fatal("imagePartFromFile() returned nil part or InlineData")
		}
		if part.InlineData.MIMEType != "image/png" {
			t.Errorf("MIMEType = %q, want %q", part.InlineData.MIMEType, "image/png")
		}
	})

	t.Run("file not found", func(t *testing.T) {
		t.Parallel()
		_, err := imagePartFromFile("/nonexistent/path/to/image.png")
		if err == nil {
			t.Error("imagePartFromFile() expected error for nonexistent file")
		}
	})

	t.Run("non-image file", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "test.txt")
		if err := os.WriteFile(tmpFile, []byte("hello world"), 0o644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		_, err := imagePartFromFile(tmpFile)
		if err == nil {
			t.Error("imagePartFromFile() expected error for non-image file")
		}
	})
}

func TestSaveInlineImages(t *testing.T) {
	t.Parallel()

	t.Run("save single image", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()

		resp := &genai.GenerateContentResponse{
			Candidates: []*genai.Candidate{
				{
					Content: &genai.Content{
						Parts: []*genai.Part{
							{
								InlineData: &genai.Blob{
									MIMEType: "image/png",
									Data:     []byte{0x89, 0x50, 0x4E, 0x47}, // PNG magic bytes
								},
							},
						},
					},
				},
			},
		}

		saved, err := saveInlineImages(resp, tmpDir, "test")
		if err != nil {
			t.Fatalf("saveInlineImages() error = %v", err)
		}
		if len(saved) != 1 {
			t.Fatalf("saveInlineImages() saved %d files, want 1", len(saved))
		}

		// Verify file exists
		if _, err := os.Stat(saved[0]); os.IsNotExist(err) {
			t.Errorf("saved file does not exist: %s", saved[0])
		}

		// Verify file extension
		if filepath.Ext(saved[0]) != ".png" {
			t.Errorf("saved file extension = %q, want %q", filepath.Ext(saved[0]), ".png")
		}
	})

	t.Run("save multiple images", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()

		resp := &genai.GenerateContentResponse{
			Candidates: []*genai.Candidate{
				{
					Content: &genai.Content{
						Parts: []*genai.Part{
							{
								InlineData: &genai.Blob{
									MIMEType: "image/png",
									Data:     []byte{0x89, 0x50, 0x4E, 0x47},
								},
							},
							{
								InlineData: &genai.Blob{
									MIMEType: "image/jpeg",
									Data:     []byte{0xFF, 0xD8, 0xFF, 0xE0},
								},
							},
						},
					},
				},
			},
		}

		saved, err := saveInlineImages(resp, tmpDir, "multi")
		if err != nil {
			t.Fatalf("saveInlineImages() error = %v", err)
		}
		if len(saved) != 2 {
			t.Fatalf("saveInlineImages() saved %d files, want 2", len(saved))
		}
	})

	t.Run("nil response", func(t *testing.T) {
		t.Parallel()
		_, err := saveInlineImages(nil, t.TempDir(), "test")
		if err == nil {
			t.Error("saveInlineImages() expected error for nil response")
		}
	})

	t.Run("empty response", func(t *testing.T) {
		t.Parallel()
		resp := &genai.GenerateContentResponse{
			Candidates: []*genai.Candidate{},
		}

		_, err := saveInlineImages(resp, t.TempDir(), "test")
		if err == nil {
			t.Error("saveInlineImages() expected error for empty response")
		}
	})

	t.Run("default dir and prefix", func(t *testing.T) {
		// Use temp dir as working directory
		originalWd, _ := os.Getwd()
		tmpDir := t.TempDir()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("failed to chdir to tmpDir: %v", err)
		}
		defer func() { _ = os.Chdir(originalWd) }()

		resp := &genai.GenerateContentResponse{
			Candidates: []*genai.Candidate{
				{
					Content: &genai.Content{
						Parts: []*genai.Part{
							{
								InlineData: &genai.Blob{
									MIMEType: "image/png",
									Data:     []byte{0x89, 0x50, 0x4E, 0x47},
								},
							},
						},
					},
				},
			},
		}

		saved, err := saveInlineImages(resp, "", "")
		if err != nil {
			t.Fatalf("saveInlineImages() error = %v", err)
		}

		// Should use default "dist" directory and "generated" prefix
		if !strings.HasPrefix(filepath.Base(saved[0]), "generated-") {
			t.Errorf("expected prefix 'generated-', got %s", filepath.Base(saved[0]))
		}
	})
}
