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

// EditEntry represents an edit entry (edit-meta.yaml)
type EditEntry struct {
	ID         string         `yaml:"id"`
	CreatedAt  string         `yaml:"created_at"`
	Source     EditSource     `yaml:"source"`
	Generation EditGeneration `yaml:"generation"`
	Result     Result         `yaml:"result"`
}

// EditGeneration contains edit generation parameters
type EditGeneration struct {
	PromptFile  string `yaml:"prompt_file"`
	AspectRatio string `yaml:"aspect_ratio,omitempty"`
	ImageSize   string `yaml:"image_size,omitempty"`
}

// EditSource contains information about the source of the edit
type EditSource struct {
	Type   string `yaml:"type"`              // "generate" or "edit"
	EditID string `yaml:"edit_id,omitempty"` // edit ID if type is "edit"
	Output string `yaml:"output"`            // source output image filename
}

const (
	editMetaFile   = "edit-meta.yaml"
	EditPromptFile = "edit-prompt.txt"
	editsDir       = "edits"
)

// NewEditEntry creates a new edit entry with a UUID v7 ID
func NewEditEntry() *EditEntry {
	id := uuid.Must(uuid.NewV7())
	return &EditEntry{
		ID:        id.String(),
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		Generation: EditGeneration{
			PromptFile: EditPromptFile,
		},
	}
}

// GetEditsDir returns the path to the edits directory within an entry
func GetEditsDir(entryDir string) string {
	return filepath.Join(entryDir, editsDir)
}

// GetEditEntryDir returns the path to a specific edit entry directory
func (e *EditEntry) GetEditEntryDir(entryDir string) string {
	return filepath.Join(GetEditsDir(entryDir), e.ID)
}

// Save writes the edit entry to the edits directory
func (e *EditEntry) Save(entryDir string) error {
	editDir := e.GetEditEntryDir(entryDir)
	if err := os.MkdirAll(editDir, 0o755); err != nil {
		return fmt.Errorf("failed to create edit directory: %w", err)
	}

	data, err := yaml.Marshal(e)
	if err != nil {
		return fmt.Errorf("failed to marshal edit entry: %w", err)
	}

	metaPath := filepath.Join(editDir, editMetaFile)
	if err := os.WriteFile(metaPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write edit-meta.yaml: %w", err)
	}

	return nil
}

// SavePrompt saves the edit prompt text to the edit entry directory
func (e *EditEntry) SavePrompt(entryDir, prompt string) error {
	editDir := e.GetEditEntryDir(entryDir)
	promptPath := filepath.Join(editDir, EditPromptFile)
	if err := os.WriteFile(promptPath, []byte(prompt), 0o644); err != nil {
		return fmt.Errorf("failed to write edit-prompt.txt: %w", err)
	}
	return nil
}

// Cleanup removes the edit entry directory (use on edit failure)
func (e *EditEntry) Cleanup(entryDir string) error {
	editDir := e.GetEditEntryDir(entryDir)
	return os.RemoveAll(editDir)
}

// loadEditEntry reads an edit entry from the specified directory
func loadEditEntry(editDir string) (*EditEntry, error) {
	metaPath := filepath.Join(editDir, editMetaFile)
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read edit-meta.yaml: %w", err)
	}

	var entry EditEntry
	if err := yaml.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("failed to parse edit-meta.yaml: %w", err)
	}

	return &entry, nil
}

// LoadEditPrompt reads the edit prompt from the edit entry directory
func LoadEditPrompt(editDir string) (string, error) {
	promptPath := filepath.Join(editDir, EditPromptFile)
	data, err := os.ReadFile(promptPath)
	if err != nil {
		return "", fmt.Errorf("failed to read edit-prompt.txt: %w", err)
	}
	return string(data), nil
}

// ListEditEntries returns all edit entries in the entry directory, sorted by UUID v7 (chronological)
func ListEditEntries(entryDir string) ([]*EditEntry, error) {
	editsPath := GetEditsDir(entryDir)
	entries, err := os.ReadDir(editsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []*EditEntry{}, nil
		}
		return nil, fmt.Errorf("failed to read edits directory: %w", err)
	}

	var result []*EditEntry
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		// Validate UUID v7 format
		if _, err := uuid.Parse(e.Name()); err != nil {
			continue
		}

		editDir := filepath.Join(editsPath, e.Name())
		entry, err := loadEditEntry(editDir)
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

// CountEditEntries returns the number of edit entries for a history entry
func CountEditEntries(entryDir string) int {
	entries, err := ListEditEntries(entryDir)
	if err != nil {
		return 0
	}
	return len(entries)
}

// GetLatestEditEntry returns the most recent edit entry
func GetLatestEditEntry(entryDir string) (*EditEntry, error) {
	entries, err := ListEditEntries(entryDir)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("no edit entries found")
	}
	return entries[len(entries)-1], nil
}

// GetEditEntryByID returns an edit entry by its ID
func GetEditEntryByID(entryDir, id string) (*EditEntry, error) {
	editDir := filepath.Join(GetEditsDir(entryDir), id)
	return loadEditEntry(editDir)
}

// GetEditOutputPath returns the path to an output image in an edit entry
func GetEditOutputPath(entryDir, editID, outputFilename string) string {
	return filepath.Join(GetEditsDir(entryDir), editID, outputFilename)
}
