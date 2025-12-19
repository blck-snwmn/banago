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
