package dto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageIsEmptyContent(t *testing.T) {
	tests := []struct {
		name    string
		content any
		want    bool
	}{
		{name: "nil", content: nil, want: true},
		{name: "empty string", content: "", want: true},
		{name: "whitespace string", content: "   ", want: true},
		{name: "non-empty string", content: "hi", want: false},
		{name: "empty array", content: []any{}, want: true},
		{name: "empty text part", content: []any{map[string]any{"type": "text", "text": ""}}, want: true},
		{name: "non-empty text part", content: []any{map[string]any{"type": "text", "text": "hi"}}, want: false},
		{
			name:    "image part only",
			content: []any{map[string]any{"type": "image_url", "image_url": map[string]any{"url": "data:image/png;base64,abc"}}},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Message{Content: tt.content}
			assert.Equal(t, tt.want, m.IsEmptyContent())
		})
	}
}

func TestClaudeRequestIsEmptySystem(t *testing.T) {
	tests := []struct {
		name   string
		system any
		want   bool
	}{
		{name: "nil", system: nil, want: true},
		{name: "empty string", system: "", want: true},
		{name: "whitespace string", system: "   ", want: true},
		{name: "non-empty string", system: "hi", want: false},
		{name: "empty array", system: []any{}, want: true},
		{name: "array with empty text block", system: []any{map[string]any{"type": "text", "text": ""}}, want: true},
		{name: "array with non-empty text block", system: []any{map[string]any{"type": "text", "text": "hi"}}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ClaudeRequest{System: tt.system}
			assert.Equal(t, tt.want, c.IsEmptySystem())
		})
	}
}
