package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var testBinPath string

func TestMain(m *testing.M) {
	// Build binary once before all tests
	tmpDir, err := os.MkdirTemp("", "banago-test-*")
	if err != nil {
		panic("failed to create temp dir: " + err.Error())
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	testBinPath = filepath.Join(tmpDir, "banago")
	buildCmd := exec.Command("go", "build", "-o", testBinPath, "github.com/blck-snwmn/banago")
	if output, err := buildCmd.CombinedOutput(); err != nil {
		panic("failed to build binary: " + err.Error() + "\n" + string(output))
	}

	os.Exit(m.Run())
}

func TestHelpCommandWithoutAPIKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
	}{
		{"root help flag", []string{"--help"}},
		{"root help shorthand", []string{"-h"}},
		{"help subcommand", []string{"help"}},
		{"generate help flag", []string{"generate", "--help"}},
		{"generate help shorthand", []string{"generate", "-h"}},
		{"help generate subcommand", []string{"help", "generate"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := exec.Command(testBinPath, tt.args...)
			// Ensure no API key is set
			cmd.Env = filterEnv(os.Environ(), "GEMINI_API_KEY")

			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Errorf("help command failed with args %v: %v\noutput: %s", tt.args, err, output)
			}
		})
	}
}

func TestGenerateCommandRequiresAPIKey(t *testing.T) {
	t.Parallel()

	// Create a dummy image file
	tmpDir := t.TempDir()
	dummyImage := filepath.Join(tmpDir, "test.png")
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
	if err := os.WriteFile(dummyImage, pngData, 0o644); err != nil {
		t.Fatalf("failed to create dummy image: %v", err)
	}

	cmd := exec.Command(testBinPath, "generate", "--prompt", "test", "--image", dummyImage)
	// Ensure no API key is set
	cmd.Env = filterEnv(os.Environ(), "GEMINI_API_KEY")

	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Error("generate command should fail without API key")
	}

	// Verify the error message mentions API key
	if !strings.Contains(string(output), "API キー") {
		t.Errorf("expected error message about API key, got: %s", output)
	}
}

// filterEnv returns a copy of env with the specified key removed
func filterEnv(env []string, key string) []string {
	result := make([]string, 0, len(env))
	prefix := key + "="
	for _, e := range env {
		if !strings.HasPrefix(e, prefix) {
			result = append(result, e)
		}
	}
	return result
}
