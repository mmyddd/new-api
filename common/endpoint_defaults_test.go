package common

import (
	"testing"

	"github.com/QuantumNous/new-api/constant"
	"github.com/stretchr/testify/assert"
)

func TestPath2EndpointType(t *testing.T) {
	tests := []struct {
		name string
		path string
		want constant.EndpointType
	}{
		{"openai chat", "/v1/chat/completions", constant.EndpointTypeOpenAI},
		{"openai chat with query", "/v1/chat/completions?stream=true", constant.EndpointTypeOpenAI},
		{"responses compact before responses", "/v1/responses/compact", constant.EndpointTypeOpenAIResponseCompact},
		{"responses", "/v1/responses", constant.EndpointTypeOpenAIResponse},
		{"alpha search", "/v1/alpha/search", constant.EndpointTypeOpenAIAlphaSearch},
		{"anthropic messages", "/v1/messages", constant.EndpointTypeAnthropic},
		{"gemini generate", "/v1beta/models/gemini-2.0-flash:generateContent", constant.EndpointTypeGemini},
		{"gemini stream", "/v1beta/models/gemini-2.0-flash:streamGenerateContent?alt=sse", constant.EndpointTypeGemini},
		{"rerank", "/v1/rerank", constant.EndpointTypeJinaRerank},
		{"image generation", "/v1/images/generations", constant.EndpointTypeImageGeneration},
		{"embeddings", "/v1/embeddings", constant.EndpointTypeEmbeddings},
		{"openai video", "/v1/video/generations", constant.EndpointTypeOpenAIVideo},
		{"models list is not gemini", "/v1/models", ""},
		{"audio has no endpoint type", "/v1/audio/speech", ""},
		{"unknown path", "/v1/foo", ""},
		{"empty path", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Path2EndpointType(tt.path))
		})
	}
}

func TestNormalizeBaseURL(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"plain domain", "https://api.openai.com", "https://api.openai.com"},
		{"v1 suffix", "https://api.openai.com/v1", "https://api.openai.com"},
		{"v1 suffix with trailing slash", "https://api.openai.com/v1/", "https://api.openai.com"},
		{"v1beta untouched", "https://generativelanguage.googleapis.com/v1beta", "https://generativelanguage.googleapis.com/v1beta"},
		{"path segment not v1", "https://x.com/api/v1/chat", "https://x.com/api/v1/chat"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NormalizeBaseURL(tt.in))
		})
	}
}
