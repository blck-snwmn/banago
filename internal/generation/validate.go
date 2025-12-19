package generation

import (
	"errors"
	"fmt"
	"os"
	"regexp"
)

// Valid image sizes
var validSizes = map[string]bool{
	"1K": true,
	"2K": true,
	"4K": true,
}

// aspectRatioRegex matches patterns like "1:1", "16:9", "4:3"
var aspectRatioRegex = regexp.MustCompile(`^\d+:\d+$`)

// validateAspectRatio validates the aspect ratio format (N:N pattern).
// Empty string is allowed (uses API default).
func validateAspectRatio(aspect string) error {
	if aspect == "" {
		return nil
	}
	if !aspectRatioRegex.MatchString(aspect) {
		return fmt.Errorf("invalid aspect ratio %q: must be in N:N format (e.g., 1:1, 16:9)", aspect)
	}
	return nil
}

// validateImageSize validates the image size value.
// Empty string is allowed (uses API default).
func validateImageSize(size string) error {
	if size == "" {
		return nil
	}
	if !validSizes[size] {
		return fmt.Errorf("invalid image size %q: must be 1K, 2K, or 4K", size)
	}
	return nil
}

// validateInputImages checks that all input image files exist.
func validateInputImages(paths []string) error {
	var missing []string
	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			missing = append(missing, path)
		}
	}
	if len(missing) > 0 {
		if len(missing) == 1 {
			return fmt.Errorf("input image not found: %s", missing[0])
		}
		return fmt.Errorf("input images not found: %v", missing)
	}
	return nil
}

// validateContext validates the generation context before making API calls.
func validateContext(ctx *Context) error {
	if err := validateAspectRatio(ctx.AspectRatio); err != nil {
		return err
	}
	if err := validateImageSize(ctx.ImageSize); err != nil {
		return err
	}
	if len(ctx.ImagePaths) == 0 {
		return errors.New("no input images specified")
	}
	if err := validateInputImages(ctx.ImagePaths); err != nil {
		return err
	}
	return nil
}
