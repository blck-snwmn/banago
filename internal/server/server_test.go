package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/blck-snwmn/banago/internal/config"
	"github.com/blck-snwmn/banago/internal/history"
)

func setupTestProject(t *testing.T) string {
	t.Helper()

	projectRoot := t.TempDir()

	// Create project config
	cfg := config.NewProjectConfig("test-project")
	if err := cfg.Save(projectRoot); err != nil {
		t.Fatalf("failed to save project config: %v", err)
	}

	// Create subproject
	subprojectDir := config.GetSubprojectDir(projectRoot, "test-subproject")
	if err := os.MkdirAll(subprojectDir, 0o755); err != nil {
		t.Fatalf("failed to create subproject dir: %v", err)
	}

	subCfg := config.NewSubprojectConfig("test-subproject")
	if err := subCfg.Save(subprojectDir); err != nil {
		t.Fatalf("failed to save subproject config: %v", err)
	}

	// Create history entry with image
	historyDir := config.GetHistoryDir(subprojectDir)
	entryDir := filepath.Join(historyDir, "test-entry-id")
	if err := os.MkdirAll(entryDir, 0o755); err != nil {
		t.Fatalf("failed to create entry dir: %v", err)
	}

	// Create a test image file
	testImagePath := filepath.Join(entryDir, "output.png")
	if err := os.WriteFile(testImagePath, []byte("fake-png-data"), 0o644); err != nil {
		t.Fatalf("failed to write test image: %v", err)
	}

	// Create meta.yaml
	entry := &history.Entry{
		ID:        "test-entry-id",
		CreatedAt: "2024-01-01T00:00:00Z",
		Result: history.Result{
			Success:      true,
			OutputImages: []string{"output.png"},
		},
	}
	if err := entry.Save(historyDir); err != nil {
		t.Fatalf("failed to save entry: %v", err)
	}

	return projectRoot
}

func TestHandleImage(t *testing.T) {
	t.Parallel()

	projectRoot := setupTestProject(t)
	srv := New(projectRoot, 8080)

	tests := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{
			name:       "valid image path",
			path:       "/images/test-subproject/test-entry-id/output.png",
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing subproject",
			path:       "/images/",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "missing entry id",
			path:       "/images/test-subproject/",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "missing filename",
			path:       "/images/test-subproject/test-entry-id/",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "only subproject",
			path:       "/images/test-subproject",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "nonexistent subproject",
			path:       "/images/nonexistent/test-entry-id/output.png",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "nonexistent file",
			path:       "/images/test-subproject/test-entry-id/nonexistent.png",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "path traversal attempt",
			path:       "/images/test-subproject/test-entry-id/../../../etc/passwd",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "path traversal with encoded dots",
			path:       "/images/test-subproject/..%2F..%2Fetc/passwd",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			srv.handleImage(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("handleImage() status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}

func TestHandleImageContent(t *testing.T) {
	t.Parallel()

	projectRoot := setupTestProject(t)
	srv := New(projectRoot, 8080)

	req := httptest.NewRequest(http.MethodGet, "/images/test-subproject/test-entry-id/output.png", nil)
	rec := httptest.NewRecorder()

	srv.handleImage(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("handleImage() status = %d, want %d", rec.Code, http.StatusOK)
	}

	body := rec.Body.String()
	if body != "fake-png-data" {
		t.Errorf("handleImage() body = %q, want %q", body, "fake-png-data")
	}
}
