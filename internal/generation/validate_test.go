package generation

import (
	"os"
	"path/filepath"
	"testing"
)

func Test_validateAspectRatio(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		aspect  string
		wantErr bool
	}{
		{"empty allowed", "", false},
		{"1:1", "1:1", false},
		{"16:9", "16:9", false},
		{"4:3", "4:3", false},
		{"9:16", "9:16", false},
		{"invalid format no colon", "169", true},
		{"invalid format text", "wide", true},
		{"invalid format partial", "16:", true},
		{"invalid format partial2", ":9", true},
		{"invalid format spaces", "16 : 9", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateAspectRatio(tt.aspect)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAspectRatio(%q) error = %v, wantErr %v", tt.aspect, err, tt.wantErr)
			}
		})
	}
}

func Test_validateImageSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		size    string
		wantErr bool
	}{
		{"empty allowed", "", false},
		{"1K valid", "1K", false},
		{"2K valid", "2K", false},
		{"4K valid", "4K", false},
		{"lowercase invalid", "1k", true},
		{"8K invalid", "8K", true},
		{"HD invalid", "HD", true},
		{"number only", "1024", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateImageSize(tt.size)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateImageSize(%q) error = %v, wantErr %v", tt.size, err, tt.wantErr)
			}
		})
	}
}

func Test_validateInputImages(t *testing.T) {
	t.Parallel()

	t.Run("empty paths allowed", func(t *testing.T) {
		t.Parallel()
		err := validateInputImages([]string{})
		if err != nil {
			t.Errorf("validateInputImages([]) error = %v, want nil", err)
		}
	})

	t.Run("existing file", func(t *testing.T) {
		t.Parallel()
		testFile := filepath.Join("..", "..", "testdata", "sample.png")
		err := validateInputImages([]string{testFile})
		if err != nil {
			t.Errorf("validateInputImages() error = %v, want nil", err)
		}
	})

	t.Run("missing single file", func(t *testing.T) {
		t.Parallel()
		err := validateInputImages([]string{"/nonexistent/image.png"})
		if err == nil {
			t.Error("validateInputImages() expected error for missing file")
		}
	})

	t.Run("missing multiple files", func(t *testing.T) {
		t.Parallel()
		err := validateInputImages([]string{"/nonexistent/a.png", "/nonexistent/b.png"})
		if err == nil {
			t.Error("validateInputImages() expected error for missing files")
		}
	})

	t.Run("mix of existing and missing", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		existingFile := filepath.Join(tmpDir, "exists.png")
		if err := os.WriteFile(existingFile, []byte{0x89, 0x50, 0x4E, 0x47}, 0o644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		err := validateInputImages([]string{existingFile, "/nonexistent/missing.png"})
		if err == nil {
			t.Error("validateInputImages() expected error when some files missing")
		}
	})
}

func Test_validateContext(t *testing.T) {
	t.Parallel()

	testFile := filepath.Join("..", "..", "testdata", "sample.png")

	t.Run("valid context", func(t *testing.T) {
		t.Parallel()
		ctx := &Context{
			Model:       "gemini-2.0-flash-exp-image-generation",
			Prompt:      "test prompt",
			ImagePaths:  []string{testFile},
			AspectRatio: "16:9",
			ImageSize:   "2K",
		}
		err := validateContext(ctx)
		if err != nil {
			t.Errorf("validateContext() error = %v, want nil", err)
		}
	})

	t.Run("invalid aspect ratio", func(t *testing.T) {
		t.Parallel()
		ctx := &Context{
			Model:       "gemini-2.0-flash-exp-image-generation",
			Prompt:      "test prompt",
			ImagePaths:  []string{testFile},
			AspectRatio: "invalid",
			ImageSize:   "2K",
		}
		err := validateContext(ctx)
		if err == nil {
			t.Error("validateContext() expected error for invalid aspect ratio")
		}
	})

	t.Run("invalid image size", func(t *testing.T) {
		t.Parallel()
		ctx := &Context{
			Model:       "gemini-2.0-flash-exp-image-generation",
			Prompt:      "test prompt",
			ImagePaths:  []string{testFile},
			AspectRatio: "16:9",
			ImageSize:   "8K",
		}
		err := validateContext(ctx)
		if err == nil {
			t.Error("validateContext() expected error for invalid image size")
		}
	})

	t.Run("no input images", func(t *testing.T) {
		t.Parallel()
		ctx := &Context{
			Model:       "gemini-2.0-flash-exp-image-generation",
			Prompt:      "test prompt",
			ImagePaths:  []string{},
			AspectRatio: "16:9",
			ImageSize:   "2K",
		}
		err := validateContext(ctx)
		if err == nil {
			t.Error("validateContext() expected error for no input images")
		}
	})

	t.Run("missing input image", func(t *testing.T) {
		t.Parallel()
		ctx := &Context{
			Model:       "gemini-2.0-flash-exp-image-generation",
			Prompt:      "test prompt",
			ImagePaths:  []string{"/nonexistent/image.png"},
			AspectRatio: "16:9",
			ImageSize:   "2K",
		}
		err := validateContext(ctx)
		if err == nil {
			t.Error("validateContext() expected error for missing input image")
		}
	})

	t.Run("empty optional fields allowed", func(t *testing.T) {
		t.Parallel()
		ctx := &Context{
			Model:       "gemini-2.0-flash-exp-image-generation",
			Prompt:      "test prompt",
			ImagePaths:  []string{testFile},
			AspectRatio: "",
			ImageSize:   "",
		}
		err := validateContext(ctx)
		if err != nil {
			t.Errorf("validateContext() error = %v, want nil for empty optional fields", err)
		}
	})
}
