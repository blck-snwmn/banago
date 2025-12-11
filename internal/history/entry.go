package history

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

// Entry represents a history entry (meta.yaml)
type Entry struct {
	ID         string     `yaml:"id"`
	CreatedAt  string     `yaml:"created_at"`
	Generation Generation `yaml:"generation"`
	Result     Result     `yaml:"result"`
}

// Generation contains generation parameters
type Generation struct {
	PromptFile    string   `yaml:"prompt_file"`
	InputImages   []string `yaml:"input_images"`
	ContextFile   string   `yaml:"context_file,omitempty"`
	CharacterFile string   `yaml:"character_file,omitempty"`
}

// Result contains generation results
type Result struct {
	Success      bool       `yaml:"success"`
	OutputImages []string   `yaml:"output_images,omitempty"`
	TokenUsage   TokenUsage `yaml:"token_usage,omitempty"`
	ErrorMessage string     `yaml:"error_message,omitempty"`
}

// TokenUsage contains token usage information
type TokenUsage struct {
	Prompt     int `yaml:"prompt"`
	Candidates int `yaml:"candidates"`
	Total      int `yaml:"total"`
	Cached     int `yaml:"cached,omitempty"`
	Thoughts   int `yaml:"thoughts,omitempty"`
}

const (
	metaFile      = "meta.yaml"
	PromptFile    = "prompt.txt"
	ContextFile   = "context.md"
	CharacterFile = "character.md"
)

// NewEntry creates a new history entry with a UUID v7 ID
func NewEntry() *Entry {
	id := uuid.Must(uuid.NewV7())
	return &Entry{
		ID:        id.String(),
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// NewEntryFromSource creates a new entry copying Generation metadata from source
func NewEntryFromSource(source *Entry) *Entry {
	entry := NewEntry()
	entry.Generation.PromptFile = source.Generation.PromptFile
	entry.Generation.InputImages = append([]string{}, source.Generation.InputImages...)
	entry.Generation.ContextFile = source.Generation.ContextFile
	entry.Generation.CharacterFile = source.Generation.CharacterFile
	return entry
}

// Save writes the entry to the history directory
func (e *Entry) Save(historyDir string) error {
	entryDir := filepath.Join(historyDir, e.ID)
	if err := os.MkdirAll(entryDir, 0o755); err != nil {
		return fmt.Errorf("failed to create entry directory: %w", err)
	}

	data, err := yaml.Marshal(e)
	if err != nil {
		return fmt.Errorf("failed to marshal entry: %w", err)
	}

	metaPath := filepath.Join(entryDir, metaFile)
	if err := os.WriteFile(metaPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write meta.yaml: %w", err)
	}

	return nil
}

// SavePrompt saves the prompt text to the entry directory
func (e *Entry) SavePrompt(historyDir, prompt string) error {
	entryDir := filepath.Join(historyDir, e.ID)
	promptPath := filepath.Join(entryDir, PromptFile)
	if err := os.WriteFile(promptPath, []byte(prompt), 0o644); err != nil {
		return fmt.Errorf("failed to write prompt.txt: %w", err)
	}
	return nil
}

// SaveContextFile copies the context file to the entry directory
func (e *Entry) SaveContextFile(historyDir, srcPath string) error {
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read context file: %w", err)
	}
	entryDir := filepath.Join(historyDir, e.ID)
	dstPath := filepath.Join(entryDir, ContextFile)
	if err := os.WriteFile(dstPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write context.md: %w", err)
	}
	return nil
}

// SaveCharacterFile copies the character file to the entry directory
func (e *Entry) SaveCharacterFile(historyDir, srcPath string) error {
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read character file: %w", err)
	}
	entryDir := filepath.Join(historyDir, e.ID)
	dstPath := filepath.Join(entryDir, CharacterFile)
	if err := os.WriteFile(dstPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write character.md: %w", err)
	}
	return nil
}

// GetEntryDir returns the path to the entry directory
func (e *Entry) GetEntryDir(historyDir string) string {
	return filepath.Join(historyDir, e.ID)
}

// Cleanup removes the entry directory (use on generation failure)
func (e *Entry) Cleanup(historyDir string) error {
	entryDir := e.GetEntryDir(historyDir)
	return os.RemoveAll(entryDir)
}

// loadEntry reads an entry from the specified directory
func loadEntry(entryDir string) (*Entry, error) {
	metaPath := filepath.Join(entryDir, metaFile)
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read meta.yaml: %w", err)
	}

	var entry Entry
	if err := yaml.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("failed to parse meta.yaml: %w", err)
	}

	return &entry, nil
}

// LoadPrompt reads the prompt from the entry directory
func LoadPrompt(entryDir string) (string, error) {
	promptPath := filepath.Join(entryDir, PromptFile)
	data, err := os.ReadFile(promptPath)
	if err != nil {
		return "", fmt.Errorf("failed to read prompt.txt: %w", err)
	}
	return string(data), nil
}

// ListEntries returns all entries in the history directory, sorted by UUID v7 (chronological)
func ListEntries(historyDir string) ([]*Entry, error) {
	entries, err := os.ReadDir(historyDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*Entry{}, nil
		}
		return nil, fmt.Errorf("failed to read history directory: %w", err)
	}

	var result []*Entry
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		// Validate UUID v7 format
		if _, err := uuid.Parse(e.Name()); err != nil {
			continue
		}

		entryDir := filepath.Join(historyDir, e.Name())
		entry, err := loadEntry(entryDir)
		if err != nil {
			continue // Skip invalid entries
		}
		result = append(result, entry)
	}

	// Sort by UUID v7 (which is chronologically sortable)
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})

	return result, nil
}

// GetLatestEntry returns the most recent entry
func GetLatestEntry(historyDir string) (*Entry, error) {
	entries, err := ListEntries(historyDir)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("no history entries found")
	}
	return entries[len(entries)-1], nil
}

// GetEntryByID returns an entry by its ID
func GetEntryByID(historyDir, id string) (*Entry, error) {
	entryDir := filepath.Join(historyDir, id)
	return loadEntry(entryDir)
}
