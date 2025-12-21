package testutil

import (
	"context"
	"sync"

	"github.com/blck-snwmn/banago/internal/gemini"
	"google.golang.org/genai"
)

// MockGenerator is a mock implementation of generation.Generator for testing.
type MockGenerator struct {
	mu sync.Mutex

	// Configuration fields
	ResponseImages [][]byte          // Image data to return
	ResponseMIME   string            // MIME type for response images (default: image/png)
	TokenUsage     gemini.TokenUsage // Token usage to return
	Error          error             // Error to return (if set, overrides success response)

	// Recording fields
	Calls []gemini.Params // Records all Generate calls
}

// Generate implements generation.Generator.
func (m *MockGenerator) Generate(_ context.Context, params gemini.Params) *gemini.Result {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Calls = append(m.Calls, params)

	if m.Error != nil {
		return &gemini.Result{Error: m.Error}
	}

	mimeType := m.ResponseMIME
	if mimeType == "" {
		mimeType = "image/png"
	}

	var parts []*genai.Part
	for _, data := range m.ResponseImages {
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
				PromptTokenCount:     int32(m.TokenUsage.Prompt),
				CandidatesTokenCount: int32(m.TokenUsage.Candidates),
				TotalTokenCount:      int32(m.TokenUsage.Total),
			},
		},
		TokenUsage: m.TokenUsage,
	}
}

// NewSuccessMock creates a MockGenerator that returns a single image.
func NewSuccessMock(imageData []byte) *MockGenerator {
	return &MockGenerator{
		ResponseImages: [][]byte{imageData},
		ResponseMIME:   "image/png",
		TokenUsage: gemini.TokenUsage{
			Prompt:     100,
			Candidates: 50,
			Total:      150,
		},
	}
}

// NewMultiImageMock creates a MockGenerator that returns multiple images.
func NewMultiImageMock(imageData []byte, count int) *MockGenerator {
	images := make([][]byte, count)
	for i := range images {
		images[i] = imageData
	}
	return &MockGenerator{
		ResponseImages: images,
		ResponseMIME:   "image/png",
		TokenUsage: gemini.TokenUsage{
			Prompt:     100,
			Candidates: 100,
			Total:      200,
		},
	}
}

// NewErrorMock creates a MockGenerator that returns an error.
func NewErrorMock(err error) *MockGenerator {
	return &MockGenerator{Error: err}
}

// CallCount returns the number of times Generate was called.
func (m *MockGenerator) CallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.Calls)
}

// LastCall returns the parameters of the last Generate call.
func (m *MockGenerator) LastCall() *gemini.Params {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.Calls) == 0 {
		return nil
	}
	return &m.Calls[len(m.Calls)-1]
}
