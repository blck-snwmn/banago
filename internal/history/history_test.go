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

func TestEntry_SaveInputImages(t *testing.T) {
	t.Parallel()

	historyDir := t.TempDir()
	entry := NewEntry()

	// Create entry dir
	entryDir := entry.GetEntryDir(historyDir)
	if err := os.MkdirAll(entryDir, 0o755); err != nil {
		t.Fatalf("failed to create entry dir: %v", err)
	}

	// Create source image files
	srcDir := t.TempDir()
	imgPaths := []string{
		filepath.Join(srcDir, "img1.png"),
		filepath.Join(srcDir, "img2.jpg"),
	}
	for _, p := range imgPaths {
		if err := os.WriteFile(p, []byte("dummy image data"), 0o644); err != nil {
			t.Fatalf("failed to write source file: %v", err)
		}
	}

	if err := entry.SaveInputImages(historyDir, imgPaths); err != nil {
		t.Fatalf("SaveInputImages() error = %v", err)
	}

	// Verify files exist in entry directory
	for _, p := range imgPaths {
		dstPath := filepath.Join(entryDir, filepath.Base(p))
		if _, err := os.Stat(dstPath); os.IsNotExist(err) {
			t.Errorf("saved file does not exist: %s", dstPath)
		}
	}
}

func TestGetInputImagePaths(t *testing.T) {
	t.Parallel()

	t.Run("files exist", func(t *testing.T) {
		t.Parallel()
		entryDir := t.TempDir()
		// Create test files
		for _, name := range []string{"img1.png", "img2.jpg"} {
			if err := os.WriteFile(filepath.Join(entryDir, name), []byte("data"), 0o644); err != nil {
				t.Fatalf("failed to create file: %v", err)
			}
		}

		paths := GetInputImagePaths(entryDir, []string{"img1.png", "img2.jpg"})
		if len(paths) != 2 {
			t.Errorf("GetInputImagePaths() returned %d paths, want 2", len(paths))
		}
	})

	t.Run("files not exist", func(t *testing.T) {
		t.Parallel()
		entryDir := t.TempDir()
		paths := GetInputImagePaths(entryDir, []string{"img1.png"})
		if len(paths) != 0 {
			t.Errorf("GetInputImagePaths() returned %d paths, want 0", len(paths))
		}
	})

	t.Run("partial files exist", func(t *testing.T) {
		t.Parallel()
		entryDir := t.TempDir()
		// Create only one file
		if err := os.WriteFile(filepath.Join(entryDir, "img1.png"), []byte("data"), 0o644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}

		paths := GetInputImagePaths(entryDir, []string{"img1.png", "img2.jpg"})
		if len(paths) != 1 {
			t.Errorf("GetInputImagePaths() returned %d paths, want 1", len(paths))
		}
	})
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

// Edit entry tests

func TestNewEditEntry(t *testing.T) {
	t.Parallel()

	entry := NewEditEntry()

	if entry.ID == "" {
		t.Error("ID should not be empty")
	}
	if entry.CreatedAt == "" {
		t.Error("CreatedAt should not be empty")
	}
	if len(entry.ID) != 36 {
		t.Errorf("ID length = %d, want 36", len(entry.ID))
	}
	if entry.PromptFile != EditPromptFile {
		t.Errorf("PromptFile = %q, want %q", entry.PromptFile, EditPromptFile)
	}
}

func TestEditEntry_SaveAndLoad(t *testing.T) {
	t.Parallel()

	// Create a history entry first
	historyDir := t.TempDir()
	genEntry := NewEntry()
	if err := genEntry.Save(historyDir); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	entryDir := genEntry.GetEntryDir(historyDir)

	// Create and save edit entry
	editEntry := NewEditEntry()
	editEntry.Source = EditSource{
		Type:   "generate",
		Output: "output.png",
	}
	editEntry.Result.Success = true
	editEntry.Result.OutputImages = []string{"edited.png"}

	if err := editEntry.Save(entryDir); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load
	loaded, err := GetEditEntryByID(entryDir, editEntry.ID)
	if err != nil {
		t.Fatalf("GetEditEntryByID() error = %v", err)
	}

	if loaded.ID != editEntry.ID {
		t.Errorf("loaded.ID = %q, want %q", loaded.ID, editEntry.ID)
	}
	if loaded.Source.Type != "generate" {
		t.Errorf("loaded.Source.Type = %q, want %q", loaded.Source.Type, "generate")
	}
	if loaded.Result.Success != true {
		t.Error("loaded.Result.Success should be true")
	}
}

func TestEditEntry_SavePrompt(t *testing.T) {
	t.Parallel()

	// Create a history entry first
	historyDir := t.TempDir()
	genEntry := NewEntry()
	if err := genEntry.Save(historyDir); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	entryDir := genEntry.GetEntryDir(historyDir)

	// Create edit entry and directory
	editEntry := NewEditEntry()
	editDir := editEntry.GetEditEntryDir(entryDir)
	if err := os.MkdirAll(editDir, 0o755); err != nil {
		t.Fatalf("failed to create edit dir: %v", err)
	}

	promptText := "Change the background color"
	if err := editEntry.SavePrompt(entryDir, promptText); err != nil {
		t.Fatalf("SavePrompt() error = %v", err)
	}

	// Load and verify
	loaded, err := LoadEditPrompt(editDir)
	if err != nil {
		t.Fatalf("LoadEditPrompt() error = %v", err)
	}
	if loaded != promptText {
		t.Errorf("LoadEditPrompt() = %q, want %q", loaded, promptText)
	}
}

func TestEditEntry_Cleanup(t *testing.T) {
	t.Parallel()

	// Create a history entry first
	historyDir := t.TempDir()
	genEntry := NewEntry()
	if err := genEntry.Save(historyDir); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	entryDir := genEntry.GetEntryDir(historyDir)

	// Create and save edit entry
	editEntry := NewEditEntry()
	if err := editEntry.Save(entryDir); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	editDir := editEntry.GetEditEntryDir(entryDir)
	if _, err := os.Stat(editDir); os.IsNotExist(err) {
		t.Fatal("edit directory should exist before cleanup")
	}

	// Cleanup
	if err := editEntry.Cleanup(entryDir); err != nil {
		t.Fatalf("Cleanup() error = %v", err)
	}

	if _, err := os.Stat(editDir); !os.IsNotExist(err) {
		t.Error("edit directory should not exist after cleanup")
	}
}

func TestListEditEntries(t *testing.T) {
	t.Parallel()

	t.Run("lists entries sorted by ID", func(t *testing.T) {
		t.Parallel()

		// Create a history entry first
		historyDir := t.TempDir()
		genEntry := NewEntry()
		if err := genEntry.Save(historyDir); err != nil {
			t.Fatalf("Save() error = %v", err)
		}
		entryDir := genEntry.GetEntryDir(historyDir)

		// Create edit entries
		edit1 := NewEditEntry()
		if err := edit1.Save(entryDir); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		edit2 := NewEditEntry()
		if err := edit2.Save(entryDir); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		entries, err := ListEditEntries(entryDir)
		if err != nil {
			t.Fatalf("ListEditEntries() error = %v", err)
		}
		if len(entries) != 2 {
			t.Fatalf("ListEditEntries() returned %d entries, want 2", len(entries))
		}

		// Should be sorted by UUID (chronological)
		if entries[0].ID > entries[1].ID {
			t.Error("entries should be sorted by ID (oldest first)")
		}
	})

	t.Run("returns empty for nonexistent edits dir", func(t *testing.T) {
		t.Parallel()
		entryDir := t.TempDir()

		entries, err := ListEditEntries(entryDir)
		if err != nil {
			t.Fatalf("ListEditEntries() error = %v", err)
		}
		if len(entries) != 0 {
			t.Errorf("ListEditEntries() returned %d entries, want 0", len(entries))
		}
	})

	t.Run("ignores invalid directories", func(t *testing.T) {
		t.Parallel()

		// Create a history entry first
		historyDir := t.TempDir()
		genEntry := NewEntry()
		if err := genEntry.Save(historyDir); err != nil {
			t.Fatalf("Save() error = %v", err)
		}
		entryDir := genEntry.GetEntryDir(historyDir)

		// Create valid edit entry
		edit := NewEditEntry()
		if err := edit.Save(entryDir); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		// Create invalid directory (not UUID)
		invalidDir := filepath.Join(GetEditsDir(entryDir), "not-a-uuid")
		if err := os.MkdirAll(invalidDir, 0o755); err != nil {
			t.Fatalf("failed to create invalid dir: %v", err)
		}

		entries, err := ListEditEntries(entryDir)
		if err != nil {
			t.Fatalf("ListEditEntries() error = %v", err)
		}
		if len(entries) != 1 {
			t.Errorf("ListEditEntries() returned %d entries, want 1", len(entries))
		}
	})
}

func TestCountEditEntries(t *testing.T) {
	t.Parallel()

	// Create a history entry first
	historyDir := t.TempDir()
	genEntry := NewEntry()
	if err := genEntry.Save(historyDir); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	entryDir := genEntry.GetEntryDir(historyDir)

	// Initially 0
	if count := CountEditEntries(entryDir); count != 0 {
		t.Errorf("CountEditEntries() = %d, want 0", count)
	}

	// Add edit entries
	for i := 0; i < 3; i++ {
		edit := NewEditEntry()
		if err := edit.Save(entryDir); err != nil {
			t.Fatalf("Save() error = %v", err)
		}
	}

	if count := CountEditEntries(entryDir); count != 3 {
		t.Errorf("CountEditEntries() = %d, want 3", count)
	}
}

func TestGetLatestEditEntry(t *testing.T) {
	t.Parallel()

	t.Run("returns latest edit entry", func(t *testing.T) {
		t.Parallel()

		// Create a history entry first
		historyDir := t.TempDir()
		genEntry := NewEntry()
		if err := genEntry.Save(historyDir); err != nil {
			t.Fatalf("Save() error = %v", err)
		}
		entryDir := genEntry.GetEntryDir(historyDir)

		edit1 := NewEditEntry()
		if err := edit1.Save(entryDir); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		edit2 := NewEditEntry()
		if err := edit2.Save(entryDir); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		latest, err := GetLatestEditEntry(entryDir)
		if err != nil {
			t.Fatalf("GetLatestEditEntry() error = %v", err)
		}
		if latest.ID != edit2.ID {
			t.Errorf("GetLatestEditEntry().ID = %q, want %q", latest.ID, edit2.ID)
		}
	})

	t.Run("returns error for empty edits", func(t *testing.T) {
		t.Parallel()
		entryDir := t.TempDir()

		_, err := GetLatestEditEntry(entryDir)
		if err == nil {
			t.Error("GetLatestEditEntry() expected error for empty edits")
		}
	})
}

func TestGetEditEntryByID(t *testing.T) {
	t.Parallel()

	t.Run("finds edit entry by ID", func(t *testing.T) {
		t.Parallel()

		// Create a history entry first
		historyDir := t.TempDir()
		genEntry := NewEntry()
		if err := genEntry.Save(historyDir); err != nil {
			t.Fatalf("Save() error = %v", err)
		}
		entryDir := genEntry.GetEntryDir(historyDir)

		edit := NewEditEntry()
		edit.Result.Success = true
		if err := edit.Save(entryDir); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		found, err := GetEditEntryByID(entryDir, edit.ID)
		if err != nil {
			t.Fatalf("GetEditEntryByID() error = %v", err)
		}
		if found.ID != edit.ID {
			t.Errorf("GetEditEntryByID().ID = %q, want %q", found.ID, edit.ID)
		}
	})

	t.Run("returns error for nonexistent ID", func(t *testing.T) {
		t.Parallel()
		entryDir := t.TempDir()

		_, err := GetEditEntryByID(entryDir, "nonexistent-id")
		if err == nil {
			t.Error("GetEditEntryByID() expected error for nonexistent ID")
		}
	})
}

func TestGetEditOutputPath(t *testing.T) {
	t.Parallel()

	entryDir := "/path/to/history/entry-id"
	editID := "edit-uuid"
	outputFilename := "output.png"

	expected := "/path/to/history/entry-id/edits/edit-uuid/output.png"
	got := GetEditOutputPath(entryDir, editID, outputFilename)

	if got != expected {
		t.Errorf("GetEditOutputPath() = %q, want %q", got, expected)
	}
}

func TestGetEditsDir(t *testing.T) {
	t.Parallel()

	entryDir := "/path/to/history/entry-id"
	expected := "/path/to/history/entry-id/edits"
	got := GetEditsDir(entryDir)

	if got != expected {
		t.Errorf("GetEditsDir() = %q, want %q", got, expected)
	}
}

func TestGetHistoryDir(t *testing.T) {
	t.Parallel()

	subprojectDir := "/path/to/subproject"
	expected := "/path/to/subproject/history"
	got := GetHistoryDir(subprojectDir)

	if got != expected {
		t.Errorf("GetHistoryDir() = %q, want %q", got, expected)
	}
}

func TestLoadEditPrompt_NotFound(t *testing.T) {
	t.Parallel()

	editDir := t.TempDir()
	_, err := LoadEditPrompt(editDir)
	if err == nil {
		t.Error("LoadEditPrompt() expected error for missing file")
	}
}
