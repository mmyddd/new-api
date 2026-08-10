package relay

import (
	"encoding/json"
	"math"
	"net/http/httptest"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/relaykit/dto"
	relaytypes "github.com/QuantumNous/new-api/relaykit/types"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/types"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsResponsesEventStreamContentType(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		want        bool
	}{
		{name: "plain", contentType: "text/event-stream", want: true},
		{name: "mixed case with charset", contentType: "Text/Event-Stream; charset=utf-8", want: true},
		{name: "json", contentType: "application/json", want: false},
		{name: "empty", contentType: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isResponsesEventStreamContentType(tt.contentType))
		})
	}
}

func TestRecalcQuotaFromRatiosIgnoresInvalidMultipliers(t *testing.T) {
	info := &relaycommon.RelayInfo{
		PriceData: types.PriceData{
			Quota: 100,
		},
	}
	info.PriceData.AddOtherRatio("duration", 2)

	quota, ok := recalcQuotaFromRatios(info, map[string]float64{
		"duration": 3,
		"zero":     0,
		"negative": -1,
		"nan":      math.NaN(),
		"inf":      math.Inf(1),
	})

	require.True(t, ok)
	assert.Equal(t, 150, quota)
	assert.True(t, info.PriceData.HasOtherRatio("duration"))
}

func TestRecalcQuotaFromRatiosRejectsAllInvalidAdjustedRatios(t *testing.T) {
	info := &relaycommon.RelayInfo{
		PriceData: types.PriceData{
			Quota: 100,
		},
	}
	info.PriceData.AddOtherRatio("duration", 2)

	quota, ok := recalcQuotaFromRatios(info, map[string]float64{
		"zero":     0,
		"negative": -1,
		"nan":      math.NaN(),
		"inf":      math.Inf(1),
	})

	require.False(t, ok)
	assert.Equal(t, 0, quota)
	assert.True(t, info.PriceData.HasOtherRatio("duration"))
}

func TestApplySystemPromptIfNeeded(t *testing.T) {
	const channelPrompt = "CHANNEL PROMPT"

	tests := []struct {
		name            string
		messages        []dto.Message
		override        bool
		want            []dto.Message
		wantOverrideKey bool
	}{
		{
			name:            "no system message injects at front",
			messages:        []dto.Message{{Role: "user", Content: "hi"}},
			want:            []dto.Message{{Role: "system", Content: channelPrompt}, {Role: "user", Content: "hi"}},
			wantOverrideKey: false,
		},
		{
			name:            "empty system message is replaced",
			messages:        []dto.Message{{Role: "system", Content: ""}, {Role: "user", Content: "hi"}},
			want:            []dto.Message{{Role: "system", Content: channelPrompt}, {Role: "user", Content: "hi"}},
			wantOverrideKey: false,
		},
		{
			name:            "whitespace system message is replaced",
			messages:        []dto.Message{{Role: "system", Content: "   "}, {Role: "user", Content: "hi"}},
			want:            []dto.Message{{Role: "system", Content: channelPrompt}, {Role: "user", Content: "hi"}},
			wantOverrideKey: false,
		},
		{
			name: "system message with empty text part is replaced",
			messages: []dto.Message{
				{Role: "system", Content: []any{map[string]any{"type": "text", "text": ""}}},
				{Role: "user", Content: "hi"},
			},
			want:            []dto.Message{{Role: "system", Content: channelPrompt}, {Role: "user", Content: "hi"}},
			wantOverrideKey: false,
		},
		{
			name:            "non-empty system message keeps user prompt",
			messages:        []dto.Message{{Role: "system", Content: "USER SYS"}, {Role: "user", Content: "hi"}},
			want:            []dto.Message{{Role: "system", Content: "USER SYS"}, {Role: "user", Content: "hi"}},
			wantOverrideKey: false,
		},
		{
			name:            "non-empty system message concatenated with override",
			messages:        []dto.Message{{Role: "system", Content: "USER SYS"}, {Role: "user", Content: "hi"}},
			override:        true,
			want:            []dto.Message{{Role: "system", Content: channelPrompt + "\nUSER SYS"}, {Role: "user", Content: "hi"}},
			wantOverrideKey: true,
		},
		{
			name:            "empty system message with override is replaced",
			messages:        []dto.Message{{Role: "system", Content: ""}, {Role: "user", Content: "hi"}},
			override:        true,
			want:            []dto.Message{{Role: "system", Content: channelPrompt}, {Role: "user", Content: "hi"}},
			wantOverrideKey: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			info := &relaycommon.RelayInfo{
				ChannelMeta: &relaycommon.ChannelMeta{
					ChannelSetting: dto.ChannelSettings{
						SystemPrompt:         channelPrompt,
						SystemPromptOverride: tt.override,
					},
				},
			}
			request := &dto.GeneralOpenAIRequest{Messages: tt.messages}
			applySystemPromptIfNeeded(c, info, request)
			require.Equal(t, tt.want, request.Messages)
			assert.Equal(t, tt.wantOverrideKey, common.GetContextKeyBool(c, constant.ContextKeySystemPromptOverride))
		})
	}
}

func TestSystemPromptSurvivesOpenAIToClaudeConversion(t *testing.T) {
	// 回归：Anthropic 渠道收到 OpenAI 格式请求（/v1/chat/completions）时，
	// 渠道系统提示词必须在 chat->claude 转换后仍出现在 Claude 的 system 字段，
	// 而不是在转换后被静默丢弃（旧的注入只处理 GeneralOpenAIRequest）。
	const channelPrompt = "CHANNEL PROMPT"

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	info := &relaycommon.RelayInfo{
		ChannelMeta: &relaycommon.ChannelMeta{
			ChannelSetting: dto.ChannelSettings{SystemPrompt: channelPrompt},
		},
	}
	request := &dto.GeneralOpenAIRequest{
		Model:    "claude-3-5-sonnet",
		Messages: []dto.Message{{Role: "user", Content: "hi"}},
	}
	applySystemPromptIfNeeded(c, info, request)

	result, err := service.ConvertRequest(c, info, relaytypes.RelayFormatClaude, request)
	require.NoError(t, err)
	claudeReq, ok := result.Value.(*dto.ClaudeRequest)
	require.True(t, ok, "expected *dto.ClaudeRequest, got %T", result.Value)
	require.False(t, claudeReq.IsEmptySystem())
	systemContents := claudeReq.ParseSystem()
	require.Len(t, systemContents, 1)
	require.NotNil(t, systemContents[0].Text)
	assert.Equal(t, channelPrompt, *systemContents[0].Text)
}

func TestApplySystemPromptToResponses(t *testing.T) {
	const channelPrompt = "CHANNEL PROMPT"

	tests := []struct {
		name            string
		instructions    json.RawMessage
		override        bool
		want            string
		wantOverrideKey bool
	}{
		{name: "absent instructions are injected", instructions: nil, want: `"CHANNEL PROMPT"`},
		{name: "empty string instructions are injected", instructions: json.RawMessage(`""`), want: `"CHANNEL PROMPT"`},
		{name: "whitespace instructions are injected", instructions: json.RawMessage(`"   "`), want: `"CHANNEL PROMPT"`},
		{name: "null instructions are injected", instructions: json.RawMessage(`null`), want: `"CHANNEL PROMPT"`},
		{name: "non-empty instructions keep user prompt", instructions: json.RawMessage(`"USER SYS"`), want: `"USER SYS"`},
		{
			name:            "non-empty instructions concatenated with override",
			instructions:    json.RawMessage(`"USER SYS"`),
			override:        true,
			want:            `"CHANNEL PROMPT\nUSER SYS"`,
			wantOverrideKey: true,
		},
		{name: "empty instructions with override are replaced", instructions: json.RawMessage(`""`), override: true, want: `"CHANNEL PROMPT"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			info := &relaycommon.RelayInfo{
				ChannelMeta: &relaycommon.ChannelMeta{
					ChannelSetting: dto.ChannelSettings{
						SystemPrompt:         channelPrompt,
						SystemPromptOverride: tt.override,
					},
				},
			}
			request := &dto.OpenAIResponsesRequest{Instructions: tt.instructions}
			applySystemPromptToResponses(c, info, request)
			assert.Equal(t, tt.want, string(request.Instructions))
			assert.Equal(t, tt.wantOverrideKey, common.GetContextKeyBool(c, constant.ContextKeySystemPromptOverride))
		})
	}
}
