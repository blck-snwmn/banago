package history

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewEntry(t *testing.T) {
	t.Parallel()

	entry := NewEntry()

	if entry.ID == "" {
		t.Error("ID should not be empty")
	}
	if entry.CreatedAt == "" {
		t.Error("CreatedAt should not be empty")
	}
	// UUID v7 is 36 characters
	if len(entry.ID) != 36 {
		t.Errorf("ID length = %d, want 36", len(entry.ID))
	}
}

func TestNewEntryFromSource(t *testing.T) {
	t.Parallel()

	source := NewEntry()
	source.Generation.PromptFile = "prompt.txt"
	source.Generation.InputImages = []string{"img1.png", "img2.jpg"}
	source.Generation.ContextFile = "context.md"
	source.Generation.CharacterFile = "char.md"

	entry := NewEntryFromSource(source)

	if entry.ID == source.ID {
		t.Error("new entry should have different ID")
	}
	if entry.Generation.PromptFile != source.Generation.PromptFile {
		t.Errorf("PromptFile = %q, want %q", entry.Generation.PromptFile, source.Generation.PromptFile)
	}
	if len(entry.Generation.InputImages) != len(source.Generation.InputImages) {
		t.Errorf("InputImages length = %d, want %d", len(entry.Generation.InputImages), len(source.Generation.InputImages))
	}
	if entry.Generation.ContextFile != source.Generation.ContextFile {
		t.Errorf("ContextFile = %q, want %q", entry.Generation.ContextFile, source.Generation.ContextFile)
	}
	if entry.Generation.CharacterFile != source.Generation.CharacterFile {
		t.Errorf("CharacterFile = %q, want %q", entry.Generation.CharacterFile, source.Generation.CharacterFile)
	}

	// Verify InputImages is a copy, not the same slice
	source.Generation.InputImages[0] = "modified.png"
	if entry.Generation.InputImages[0] == "modified.png" {
		t.Error("InputImages should be a copy, not the same slice")
	}
}

func TestEntry_SaveAndLoad(t *testing.T) {
	t.Parallel()

	historyDir := t.TempDir()
	entry := NewEntry()
	entry.Generation.PromptFile = PromptFile
	entry.Generation.InputImages = []string{"img1.png"}
	entry.Result.Success = true
	entry.Result.OutputImages = []string{"output.png"}

	// Save
	if err := entry.Save(historyDir); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load
	loaded, err := GetEntryByID(historyDir, entry.ID)
	if err != nil {
		t.Fatalf("GetEntryByID() error = %v", err)
	}

	if loaded.ID != entry.ID {
		t.Errorf("loaded.ID = %q, want %q", loaded.ID, entry.ID)
	}
	if loaded.Result.Success != entry.Result.Success {
		t.Errorf("loaded.Result.Success = %v, want %v", loaded.Result.Success, entry.Result.Success)
	}
}

func TestEntry_SavePrompt(t *testing.T) {
	t.Parallel()

	historyDir := t.TempDir()
	entry := NewEntry()

	// Create entry dir
	entryDir := entry.GetEntryDir(historyDir)
	if err := os.MkdirAll(entryDir, 0o755); err != nil {
		t.Fatalf("failed to create entry dir: %v", err)
	}

	promptText := "Generate a beautiful sunset"
	if err := entry.SavePrompt(historyDir, promptText); err != nil {
		t.Fatalf("SavePrompt() error = %v", err)
	}

	// Load and verify
	loaded, err := LoadPrompt(entryDir)
	if err != nil {
		t.Fatalf("LoadPrompt() error = %v", err)
	}
	if loaded != promptText {
		t.Errorf("LoadPrompt() = %q, want %q", loaded, promptText)
	}
}

func TestEntry_SaveContextFile(t *testing.T) {
	t.Parallel()

	historyDir := t.TempDir()
	entry := NewEntry()

	// Create entry dir
	entryDir := entry.GetEntryDir(historyDir)
	if err := os.MkdirAll(entryDir, 0o755); err != nil {
		t.Fatalf("failed to create entry dir: %v", err)
	}

	// Create source context file
	srcDir := t.TempDir()
	srcPath := filepath.Join(srcDir, "context.md")
	contextContent := "# Context\nSome context info"
	if err := os.WriteFile(srcPath, []byte(contextContent), 0o644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	if err := entry.SaveContextFile(historyDir, srcPath); err != nil {
		t.Fatalf("SaveContextFile() error = %v", err)
	}

	// Verify
	dstPath := filepath.Join(entryDir, ContextFile)
	data, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}
	if string(data) != contextContent {
		t.Errorf("saved content = %q, want %q", string(data), contextContent)
	}
}

func TestEntry_SaveCharacterFile(t *testing.T) {
	t.Parallel()

	historyDir := t.TempDir()
	entry := NewEntry()

	// Create entry dir
	entryDir := entry.GetEntryDir(historyDir)
	if err := os.MkdirAll(entryDir, 0o755); err != nil {
		t.Fatalf("failed to create entry dir: %v", err)
	}

	// Create source character file
	srcDir := t.TempDir()
	srcPath := filepath.Join(srcDir, "char.md")
	charContent := "# Character\nBlue hair, red eyes"
	if err := os.WriteFile(srcPath, []byte(charContent), 0o644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	if err := entry.SaveCharacterFile(historyDir, srcPath); err != nil {
		t.Fatalf("SaveCharacterFile() error = %v", err)
	}

	// Verify
	dstPath := filepath.Join(entryDir, CharacterFile)
	data, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}
	if string(data) != charContent {
		t.Errorf("saved content = %q, want %q", string(data), charContent)
	}
}

func TestEntry_Cleanup(t *testing.T) {
	t.Parallel()

	historyDir := t.TempDir()
	entry := NewEntry()

	// Save entry
	if err := entry.Save(historyDir); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	entryDir := entry.GetEntryDir(historyDir)
	if _, err := os.Stat(entryDir); os.IsNotExist(err) {
		t.Fatal("entry directory should exist before cleanup")
	}

	// Cleanup
	if err := entry.Cleanup(historyDir); err != nil {
		t.Fatalf("Cleanup() error = %v", err)
	}

	if _, err := os.Stat(entryDir); !os.IsNotExist(err) {
		t.Error("entry directory should not exist after cleanup")
	}
}

func TestListEntries(t *testing.T) {
	t.Parallel()

	t.Run("lists entries sorted by ID", func(t *testing.T) {
		t.Parallel()
		historyDir := t.TempDir()

		// Create entries
		entry1 := NewEntry()
		entry1.Result.Success = true
		if err := entry1.Save(historyDir); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		entry2 := NewEntry()
		entry2.Result.Success = true
		if err := entry2.Save(historyDir); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		entries, err := ListEntries(historyDir)
		if err != nil {
			t.Fatalf("ListEntries() error = %v", err)
		}
		if len(entries) != 2 {
			t.Fatalf("ListEntries() returned %d entries, want 2", len(entries))
		}

		// Should be sorted by UUID (chronological)
		if entries[0].ID > entries[1].ID {
			t.Error("entries should be sorted by ID (oldest first)")
		}
	})

	t.Run("returns empty for nonexistent dir", func(t *testing.T) {
		t.Parallel()
		historyDir := filepath.Join(t.TempDir(), "nonexistent")

		entries, err := ListEntries(historyDir)
		if err != nil {
			t.Fatalf("ListEntries() error = %v", err)
		}
		if len(entries) != 0 {
			t.Errorf("ListEntries() returned %d entries, want 0", len(entries))
		}
	})

	t.Run("ignores invalid directories", func(t *testing.T) {
		t.Parallel()
		historyDir := t.TempDir()

		// Create valid entry
		entry := NewEntry()
		if err := entry.Save(historyDir); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		// Create invalid directory (not UUID)
		invalidDir := filepath.Join(historyDir, "not-a-uuid")
		if err := os.MkdirAll(invalidDir, 0o755); err != nil {
			t.Fatalf("failed to create invalid dir: %v", err)
		}

		entries, err := ListEntries(historyDir)
		if err != nil {
			t.Fatalf("ListEntries() error = %v", err)
		}
		if len(entries) != 1 {
			t.Errorf("ListEntries() returned %d entries, want 1", len(entries))
		}
	})
}

func TestGetLatestEntry(t *testing.T) {
	t.Parallel()

	t.Run("returns latest entry", func(t *testing.T) {
		t.Parallel()
		historyDir := t.TempDir()

		entry1 := NewEntry()
		if err := entry1.Save(historyDir); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		entry2 := NewEntry()
		if err := entry2.Save(historyDir); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		latest, err := GetLatestEntry(historyDir)
		if err != nil {
			t.Fatalf("GetLatestEntry() error = %v", err)
		}
		if latest.ID != entry2.ID {
			t.Errorf("GetLatestEntry().ID = %q, want %q", latest.ID, entry2.ID)
		}
	})

	t.Run("returns error for empty history", func(t *testing.T) {
		t.Parallel()
		historyDir := t.TempDir()

		_, err := GetLatestEntry(historyDir)
		if err == nil {
			t.Error("GetLatestEntry() expected error for empty history")
		}
	})
}

func TestGetEntryByID(t *testing.T) {
	t.Parallel()

	t.Run("finds entry by ID", func(t *testing.T) {
		t.Parallel()
		historyDir := t.TempDir()

		entry := NewEntry()
		entry.Result.Success = true
		if err := entry.Save(historyDir); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		found, err := GetEntryByID(historyDir, entry.ID)
		if err != nil {
			t.Fatalf("GetEntryByID() error = %v", err)
		}
		if found.ID != entry.ID {
			t.Errorf("GetEntryByID().ID = %q, want %q", found.ID, entry.ID)
		}
	})

	t.Run("returns error for nonexistent ID", func(t *testing.T) {
		t.Parallel()
		historyDir := t.TempDir()

		_, err := GetEntryByID(historyDir, "nonexistent-id")
		if err == nil {
			t.Error("GetEntryByID() expected error for nonexistent ID")
		}
	})
}

func TestLoadPrompt_NotFound(t *testing.T) {
	t.Parallel()

	entryDir := t.TempDir()
	_, err := LoadPrompt(entryDir)
	if err == nil {
		t.Error("LoadPrompt() expected error for missing file")
	}
}
