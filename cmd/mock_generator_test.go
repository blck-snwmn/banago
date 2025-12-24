package cmd

import (
	"context"
	"sync"

	"github.com/blck-snwmn/banago/internal/gemini"
	"github.com/blck-snwmn/banago/internal/generation"
	"google.golang.org/genai"
)

// Compile-time interface compliance check.
var _ generation.Generator = (*mockGenerator)(nil)

// mockGenerator is a mock implementation of generation.Generator for testing.
type mockGenerator struct {
	mu sync.Mutex

	// Configuration fields
	responseImages [][]byte          // Image data to return
	responseMIME   string            // MIME type for response images (default: image/png)
	tokenUsage     gemini.TokenUsage // Token usage to return
	err            error             // Error to return (if set, overrides success response)

	// Recording fields
	calls []gemini.Params // Records all Generate calls
}

// Generate implements generation.Generator.
func (m *mockGenerator) Generate(_ context.Context, params gemini.Params) *gemini.Result {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.calls = append(m.calls, params)

	if m.err != nil {
		return &gemini.Result{Error: m.err}
	}

	mimeType := m.responseMIME
	if mimeType == "" {
		mimeType = "image/png"
	}

	var parts []*genai.Part
	for _, data := range m.responseImages {
		parts = append(parts, &genai.Part{
			InlineData: &genai.Blob{
				MIMEType: mimeType,
				Data:     data,
			},
		})
	}

	return &gemini.Result{
		Response: &genai.GenerateContentResponse{
			Candidates: []*genai.Candidate{{
				Content: &genai.Content{Parts: parts},
			}},
			UsageMetadata: &genai.GenerateContentResponseUsageMetadata{
				PromptTokenCount:     int32(m.tokenUsage.Prompt),
				CandidatesTokenCount: int32(m.tokenUsage.Candidates),
				TotalTokenCount:      int32(m.tokenUsage.Total),
			},
		},
		TokenUsage: m.tokenUsage,
	}
}

// newSuccessMock creates a mockGenerator that returns a single image.
func newSuccessMock(imageData []byte) *mockGenerator {
	return &mockGenerator{
		responseImages: [][]byte{imageData},
		responseMIME:   "image/png",
		tokenUsage: gemini.TokenUsage{
			Prompt:     100,
			Candidates: 50,
			Total:      150,
		},
	}
}

// newMultiImageMock creates a mockGenerator that returns multiple images.
func newMultiImageMock(imageData []byte, count int) *mockGenerator {
	images := make([][]byte, count)
	for i := range images {
		images[i] = imageData
	}
	return &mockGenerator{
		responseImages: images,
		responseMIME:   "image/png",
		tokenUsage: gemini.TokenUsage{
			Prompt:     100,
			Candidates: 100,
			Total:      200,
		},
	}
}

// newErrorMock creates a mockGenerator that returns an error.
func newErrorMock(err error) *mockGenerator {
	return &mockGenerator{err: err}
}

// callCount returns the number of times Generate was called.
func (m *mockGenerator) callCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.calls)
}

// lastCall returns the parameters of the last Generate call.
func (m *mockGenerator) lastCall() *gemini.Params {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.calls) == 0 {
		return nil
	}
	return &m.calls[len(m.calls)-1]
}
