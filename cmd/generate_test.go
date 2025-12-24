package cmd

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/blck-snwmn/banago/internal/config"
	"github.com/blck-snwmn/banago/internal/gemini"
	"github.com/blck-snwmn/banago/internal/history"
	"github.com/blck-snwmn/banago/internal/project"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/genai"
)

func TestResolvePrompt(t *testing.T) {
	t.Parallel()

	t.Run("inline prompt", func(t *testing.T) {
		t.Parallel()
		got, err := resolvePrompt("hello world", "")
		if err != nil {
			t.Fatalf("resolvePrompt() error = %v", err)
		}
		if got != "hello world" {
			t.Errorf("resolvePrompt() = %q, want %q", got, "hello world")
		}
	})

	t.Run("inline prompt with whitespace", func(t *testing.T) {
		t.Parallel()
		got, err := resolvePrompt("  hello world  ", "")
		if err != nil {
			t.Fatalf("resolvePrompt() error = %v", err)
		}
		if got != "hello world" {
			t.Errorf("resolvePrompt() = %q, want %q", got, "hello world")
		}
	})

	t.Run("prompt from file", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		promptFile := filepath.Join(tmpDir, "prompt.txt")
		if err := os.WriteFile(promptFile, []byte("prompt from file"), 0o644); err != nil {
			t.Fatalf("failed to write prompt file: %v", err)
		}

		got, err := resolvePrompt("", promptFile)
		if err != nil {
			t.Fatalf("resolvePrompt() error = %v", err)
		}
		if got != "prompt from file" {
			t.Errorf("resolvePrompt() = %q, want %q", got, "prompt from file")
		}
	})

	t.Run("empty inline prompt", func(t *testing.T) {
		t.Parallel()
		_, err := resolvePrompt("   ", "")
		if err == nil {
			t.Error("resolvePrompt() expected error for empty prompt")
		}
	})

	t.Run("empty file prompt", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		promptFile := filepath.Join(tmpDir, "empty.txt")
		if err := os.WriteFile(promptFile, []byte("   "), 0o644); err != nil {
			t.Fatalf("failed to write prompt file: %v", err)
		}

		_, err := resolvePrompt("", promptFile)
		if err == nil {
			t.Error("resolvePrompt() expected error for empty file prompt")
		}
	})

	t.Run("no prompt provided", func(t *testing.T) {
		t.Parallel()
		_, err := resolvePrompt("", "")
		if err == nil {
			t.Error("resolvePrompt() expected error when no prompt provided")
		}
	})

	t.Run("file not found", func(t *testing.T) {
		t.Parallel()
		_, err := resolvePrompt("", "/nonexistent/prompt.txt")
		if err == nil {
			t.Error("resolvePrompt() expected error for nonexistent file")
		}
	})
}

func TestCollectImagePaths(t *testing.T) {
	t.Parallel()

	t.Run("from subproject config", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		subprojectDir := filepath.Join(tmpDir, "subprojects", "test")
		cfg := &config.SubprojectConfig{
			InputImages: []string{"img1.png", "img2.jpg"},
		}

		got := collectImagePaths(subprojectDir, cfg)

		inputsDir := project.GetInputsDir(subprojectDir)
		want := []string{
			filepath.Join(inputsDir, "img1.png"),
			filepath.Join(inputsDir, "img2.jpg"),
		}
		if len(got) != len(want) {
			t.Fatalf("collectImagePaths() returned %d paths, want %d", len(got), len(want))
		}
		for i, g := range got {
			if g != want[i] {
				t.Errorf("collectImagePaths()[%d] = %q, want %q", i, g, want[i])
			}
		}
	})

	t.Run("empty config returns empty", func(t *testing.T) {
		t.Parallel()
		cfg := &config.SubprojectConfig{
			InputImages: []string{},
		}

		got := collectImagePaths("", cfg)

		if len(got) != 0 {
			t.Errorf("collectImagePaths() returned %d paths, want 0", len(got))
		}
	})
}

func TestResolveGenerationParams(t *testing.T) {
	t.Parallel()

	t.Run("flags take precedence", func(t *testing.T) {
		t.Parallel()
		cfg := &config.SubprojectConfig{
			AspectRatio: "16:9",
			ImageSize:   "2K",
		}

		aspect, size := resolveGenerationParams("1:1", "4K", cfg)

		if aspect != "1:1" {
			t.Errorf("aspect = %q, want %q", aspect, "1:1")
		}
		if size != "4K" {
			t.Errorf("size = %q, want %q", size, "4K")
		}
	})

	t.Run("fallback to config", func(t *testing.T) {
		t.Parallel()
		cfg := &config.SubprojectConfig{
			AspectRatio: "16:9",
			ImageSize:   "2K",
		}

		aspect, size := resolveGenerationParams("", "", cfg)

		if aspect != "16:9" {
			t.Errorf("aspect = %q, want %q", aspect, "16:9")
		}
		if size != "2K" {
			t.Errorf("size = %q, want %q", size, "2K")
		}
	})

	t.Run("partial override", func(t *testing.T) {
		t.Parallel()
		cfg := &config.SubprojectConfig{
			AspectRatio: "16:9",
			ImageSize:   "2K",
		}

		aspect, size := resolveGenerationParams("1:1", "", cfg)

		if aspect != "1:1" {
			t.Errorf("aspect = %q, want %q", aspect, "1:1")
		}
		if size != "2K" {
			t.Errorf("size = %q, want %q", size, "2K")
		}
	})

	t.Run("empty config no flags", func(t *testing.T) {
		t.Parallel()
		cfg := &config.SubprojectConfig{}
		aspect, size := resolveGenerationParams("", "", cfg)

		if aspect != "" {
			t.Errorf("aspect = %q, want empty", aspect)
		}
		if size != "" {
			t.Errorf("size = %q, want empty", size)
		}
	})
}

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
			got := gemini.NormalizeExt(tt.mimeType)
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
		// Use shared test data
		testFile := filepath.Join("..", "testdata", "sample.png")

		part, err := gemini.ImagePartFromFile(testFile)
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
		_, err := gemini.ImagePartFromFile("/nonexistent/path/to/image.png")
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

		_, err := gemini.ImagePartFromFile(tmpFile)
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

		saved, err := gemini.SaveImages(resp, tmpDir)
		if err != nil {
			t.Fatalf("SaveImages() error = %v", err)
		}
		if len(saved) != 1 {
			t.Fatalf("SaveImages() saved %d files, want 1", len(saved))
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

		saved, err := gemini.SaveImages(resp, tmpDir)
		if err != nil {
			t.Fatalf("SaveImages() error = %v", err)
		}
		if len(saved) != 2 {
			t.Fatalf("SaveImages() saved %d files, want 2", len(saved))
		}
	})

	t.Run("nil response", func(t *testing.T) {
		t.Parallel()
		_, err := gemini.SaveImages(nil, t.TempDir())
		if err == nil {
			t.Error("SaveImages() expected error for nil response")
		}
	})

	t.Run("empty response", func(t *testing.T) {
		t.Parallel()
		resp := &genai.GenerateContentResponse{
			Candidates: []*genai.Candidate{},
		}

		_, err := gemini.SaveImages(resp, t.TempDir())
		if err == nil {
			t.Error("SaveImages() expected error for empty response")
		}
	})

}

// Handler tests - these test the full generation workflow with DI

func TestGenerateHandler_Run_Success(t *testing.T) {
	t.Parallel()

	// Setup project
	projectRoot := t.TempDir()
	require.NoError(t, project.InitProject(projectRoot, "test-project", false))
	require.NoError(t, project.CreateSubproject(projectRoot, "test-sub", ""))
	subprojectDir := project.GetSubprojectDir(projectRoot, "test-sub")

	// Configure input images
	cfg, err := config.LoadSubprojectConfig(subprojectDir)
	require.NoError(t, err)
	cfg.InputImages = []string{"test.png"}
	require.NoError(t, cfg.Save(subprojectDir))

	// Create input image file
	pngData, err := os.ReadFile("testdata/sample.png")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(project.GetInputsDir(subprojectDir), "test.png"), pngData, 0o644))

	mock := newSuccessMock(pngData)
	handler := &generateHandler{generator: mock}

	var buf bytes.Buffer
	err = handler.run(context.Background(), generateOptions{
		prompt: "test prompt",
	}, subprojectDir, &buf)
	require.NoError(t, err)

	// Verify output
	output := buf.String()
	assert.Contains(t, output, "History ID:")
	assert.Contains(t, output, "Generated files:")

	// Verify mock was called correctly
	assert.Equal(t, 1, mock.callCount())
	lastCall := mock.lastCall()
	assert.Equal(t, "gemini-3-pro-image-preview", lastCall.Model)
	assert.Equal(t, "test prompt", lastCall.Prompt)

	// Verify history entry was created
	historyDir := filepath.Join(subprojectDir, "history")
	entries, err := history.ListEntries(historyDir)
	require.NoError(t, err)
	assert.Len(t, entries, 1)

	// Verify history entry structure
	entryDir := filepath.Join(historyDir, entries[0].ID)
	assert.DirExists(t, entryDir)
	assert.FileExists(t, filepath.Join(entryDir, "meta.yaml"))
	assert.FileExists(t, filepath.Join(entryDir, "prompt.txt"))

	// Verify output images exist
	outputFiles, err := filepath.Glob(filepath.Join(entryDir, "output-*.png"))
	require.NoError(t, err)
	assert.Len(t, outputFiles, 1)
}

func TestGenerateHandler_Run_APIError(t *testing.T) {
	t.Parallel()

	// Setup project
	projectRoot := t.TempDir()
	require.NoError(t, project.InitProject(projectRoot, "test-project", false))
	require.NoError(t, project.CreateSubproject(projectRoot, "test-sub", ""))
	subprojectDir := project.GetSubprojectDir(projectRoot, "test-sub")

	// Configure input images
	cfg, err := config.LoadSubprojectConfig(subprojectDir)
	require.NoError(t, err)
	cfg.InputImages = []string{"test.png"}
	require.NoError(t, cfg.Save(subprojectDir))

	// Create input image file
	pngData, err := os.ReadFile("testdata/sample.png")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(project.GetInputsDir(subprojectDir), "test.png"), pngData, 0o644))

	apiErr := errors.New("API error")
	mock := newErrorMock(apiErr)
	handler := &generateHandler{generator: mock}

	var buf bytes.Buffer
	err = handler.run(context.Background(), generateOptions{
		prompt: "test prompt",
	}, subprojectDir, &buf)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to generate image")

	// Verify no history entries remain (cleanup)
	historyDir := filepath.Join(subprojectDir, "history")
	entries, listErr := history.ListEntries(historyDir)
	require.NoError(t, listErr)
	assert.Empty(t, entries, "expected no history entries after error")
}

func TestGenerateHandler_Run_MultipleImages(t *testing.T) {
	t.Parallel()

	// Setup project
	projectRoot := t.TempDir()
	require.NoError(t, project.InitProject(projectRoot, "test-project", false))
	require.NoError(t, project.CreateSubproject(projectRoot, "test-sub", ""))
	subprojectDir := project.GetSubprojectDir(projectRoot, "test-sub")

	// Configure input images
	cfg, err := config.LoadSubprojectConfig(subprojectDir)
	require.NoError(t, err)
	cfg.InputImages = []string{"test.png"}
	require.NoError(t, cfg.Save(subprojectDir))

	// Create input image file
	pngData, err := os.ReadFile("testdata/sample.png")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(project.GetInputsDir(subprojectDir), "test.png"), pngData, 0o644))

	mock := newMultiImageMock(pngData, 3)
	handler := &generateHandler{generator: mock}

	var buf bytes.Buffer
	err = handler.run(context.Background(), generateOptions{
		prompt: "test prompt",
	}, subprojectDir, &buf)
	require.NoError(t, err)

	// Verify 3 output images were generated
	historyDir := filepath.Join(subprojectDir, "history")
	entries, err := history.ListEntries(historyDir)
	require.NoError(t, err)
	require.Len(t, entries, 1)

	entryDir := filepath.Join(historyDir, entries[0].ID)
	outputFiles, err := filepath.Glob(filepath.Join(entryDir, "output-*.png"))
	require.NoError(t, err)
	assert.Len(t, outputFiles, 3)
}

func TestGenerateHandler_Run_ValidationError(t *testing.T) {
	t.Parallel()

	// Setup project
	projectRoot := t.TempDir()
	require.NoError(t, project.InitProject(projectRoot, "test-project", false))
	require.NoError(t, project.CreateSubproject(projectRoot, "test-sub", ""))
	subprojectDir := project.GetSubprojectDir(projectRoot, "test-sub")

	// Configure input images
	cfg, err := config.LoadSubprojectConfig(subprojectDir)
	require.NoError(t, err)
	cfg.InputImages = []string{"test.png"}
	require.NoError(t, cfg.Save(subprojectDir))

	// Create input image file
	pngData, err := os.ReadFile("testdata/sample.png")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(project.GetInputsDir(subprojectDir), "test.png"), pngData, 0o644))

	mock := newSuccessMock(pngData)
	handler := &generateHandler{generator: mock}

	t.Run("invalid aspect ratio", func(t *testing.T) {
		var buf bytes.Buffer
		err := handler.run(context.Background(), generateOptions{
			prompt: "test prompt",
			aspect: "invalid",
		}, subprojectDir, &buf)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "aspect ratio")

		// Verify mock was NOT called (validation failed before API call)
		assert.Equal(t, 0, mock.callCount())
	})

	t.Run("invalid image size", func(t *testing.T) {
		var buf bytes.Buffer
		err := handler.run(context.Background(), generateOptions{
			prompt: "test prompt",
			size:   "5K",
		}, subprojectDir, &buf)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "image size")
	})
}

func TestGenerateHandler_Run_WithAspectRatioAndSize(t *testing.T) {
	t.Parallel()

	// Setup project
	projectRoot := t.TempDir()
	require.NoError(t, project.InitProject(projectRoot, "test-project", false))
	require.NoError(t, project.CreateSubproject(projectRoot, "test-sub", ""))
	subprojectDir := project.GetSubprojectDir(projectRoot, "test-sub")

	// Configure input images
	cfg, err := config.LoadSubprojectConfig(subprojectDir)
	require.NoError(t, err)
	cfg.InputImages = []string{"test.png"}
	require.NoError(t, cfg.Save(subprojectDir))

	// Create input image file
	pngData, err := os.ReadFile("testdata/sample.png")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(project.GetInputsDir(subprojectDir), "test.png"), pngData, 0o644))

	mock := newSuccessMock(pngData)
	handler := &generateHandler{generator: mock}

	var buf bytes.Buffer
	err = handler.run(context.Background(), generateOptions{
		prompt: "test prompt",
		aspect: "16:9",
		size:   "2K",
	}, subprojectDir, &buf)
	require.NoError(t, err)

	// Verify mock received correct parameters
	lastCall := mock.lastCall()
	assert.Equal(t, "16:9", lastCall.AspectRatio)
	assert.Equal(t, "2K", lastCall.ImageSize)
}

func TestGenerateHandler_Run_PromptFromFile(t *testing.T) {
	t.Parallel()

	// Setup project
	projectRoot := t.TempDir()
	require.NoError(t, project.InitProject(projectRoot, "test-project", false))
	require.NoError(t, project.CreateSubproject(projectRoot, "test-sub", ""))
	subprojectDir := project.GetSubprojectDir(projectRoot, "test-sub")

	// Configure input images
	cfg, err := config.LoadSubprojectConfig(subprojectDir)
	require.NoError(t, err)
	cfg.InputImages = []string{"test.png"}
	require.NoError(t, cfg.Save(subprojectDir))

	// Create input image file
	pngData, err := os.ReadFile("testdata/sample.png")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(project.GetInputsDir(subprojectDir), "test.png"), pngData, 0o644))

	// Create a prompt file
	promptFile := filepath.Join(subprojectDir, "prompt.txt")
	require.NoError(t, os.WriteFile(promptFile, []byte("prompt from file"), 0o644))

	mock := newSuccessMock(pngData)
	handler := &generateHandler{generator: mock}

	var buf bytes.Buffer
	err = handler.run(context.Background(), generateOptions{
		promptFile: promptFile,
	}, subprojectDir, &buf)
	require.NoError(t, err)

	// Verify prompt was read from file
	lastCall := mock.lastCall()
	assert.Equal(t, "prompt from file", lastCall.Prompt)
}

func TestGenerateHandler_Run_ProjectNotFound(t *testing.T) {
	t.Parallel()

	// Use a temp directory without project structure
	tempDir := t.TempDir()

	pngData, err := os.ReadFile("testdata/sample.png")
	require.NoError(t, err)

	mock := newSuccessMock(pngData)
	handler := &generateHandler{generator: mock}

	var buf bytes.Buffer
	err = handler.run(context.Background(), generateOptions{
		prompt: "test prompt",
	}, tempDir, &buf)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "banago project not found")
}

func TestGenerateHandler_Run_NotInSubproject(t *testing.T) {
	t.Parallel()

	// Create project but run from project root (not subproject)
	projectRoot := t.TempDir()
	require.NoError(t, project.InitProject(projectRoot, "test-project", false))

	pngData, err := os.ReadFile("testdata/sample.png")
	require.NoError(t, err)

	mock := newSuccessMock(pngData)
	handler := &generateHandler{generator: mock}

	var buf bytes.Buffer
	err = handler.run(context.Background(), generateOptions{
		prompt: "test prompt",
	}, projectRoot, &buf)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not in a subproject")
}

// Scenario tests

func TestScenario_MultipleGenerations(t *testing.T) {
	t.Parallel()

	// Setup project
	projectRoot := t.TempDir()
	require.NoError(t, project.InitProject(projectRoot, "test-project", false))
	require.NoError(t, project.CreateSubproject(projectRoot, "test-sub", ""))
	subprojectDir := project.GetSubprojectDir(projectRoot, "test-sub")

	// Configure input images
	cfg, err := config.LoadSubprojectConfig(subprojectDir)
	require.NoError(t, err)
	cfg.InputImages = []string{"test.png"}
	require.NoError(t, cfg.Save(subprojectDir))

	// Create input image file
	pngData, err := os.ReadFile("testdata/sample.png")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(project.GetInputsDir(subprojectDir), "test.png"), pngData, 0o644))

	mock := newSuccessMock(pngData)
	handler := &generateHandler{generator: mock}

	// Generate multiple times
	for i := 0; i < 3; i++ {
		var buf bytes.Buffer
		err := handler.run(context.Background(), generateOptions{
			prompt: "test prompt",
		}, subprojectDir, &buf)
		require.NoError(t, err)
	}

	// Verify all entries are unique
	historyDir := filepath.Join(subprojectDir, "history")
	entries, err := history.ListEntries(historyDir)
	require.NoError(t, err)
	assert.Len(t, entries, 3)

	// Verify all entries have different IDs
	ids := make(map[string]bool)
	for _, entry := range entries {
		assert.False(t, ids[entry.ID], "duplicate entry ID found")
		ids[entry.ID] = true
	}

	// Verify mock was called 3 times
	assert.Equal(t, 3, mock.callCount())
}
