package generation

// Spec holds all information needed for generation and to be saved to history.
type Spec struct {
	// Generation parameters
	Model       string
	Prompt      string
	ImagePaths  []string
	AspectRatio string
	ImageSize   string

	// For history metadata - the filenames of input images
	InputImageNames []string

	// Source entry ID for regeneration tracking (empty for new generation)
	SourceEntryID string
}

// EditSpec holds all information needed for editing an existing image.
type EditSpec struct {
	// Generation parameters
	Model       string
	Prompt      string
	AspectRatio string
	ImageSize   string

	// Source image information
	SourceImagePath string

	// History context
	EntryID string // The generate entry ID

	// Source information for tracking
	SourceType   string // "generate" or "edit"
	SourceEditID string // If editing from an edit, the source edit ID
	SourceOutput string // The output filename being edited
}
