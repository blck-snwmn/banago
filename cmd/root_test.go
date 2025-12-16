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

	// API key check happens before project/subproject validation
	cmd := exec.Command(testBinPath, "generate", "--prompt", "test")
	// Ensure no API key is set
	cmd.Env = filterEnv(os.Environ(), "GEMINI_API_KEY")

	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Error("generate command should fail without API key")
	}

	// Verify the error message mentions API key
	if !strings.Contains(string(output), "API key") {
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
