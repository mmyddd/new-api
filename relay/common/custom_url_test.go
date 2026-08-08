package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveCustomChannelURL(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		requestPath string
		model       string
		want        string
	}{
		{
			// 旧数据：完整 chat URL，请求 chat → 原路径
			"full chat url + chat request",
			"https://opencode.ai/zen/go/v1/chat/completions",
			"/v1/chat/completions",
			"deepseek-v4-flash",
			"https://opencode.ai/zen/go/v1/chat/completions",
		},
		{
			// 旧数据：完整 chat URL，请求 responses → 自动换 /v1/responses
			"full chat url + responses request",
			"https://opencode.ai/zen/go/v1/chat/completions",
			"/v1/responses",
			"deepseek-v4-flash",
			"https://opencode.ai/zen/go/v1/responses",
		},
		{
			// 规范化后：/v1 结尾基础 URL + chat 请求
			"v1 base url + chat request",
			"https://opencode.ai/zen/go/v1",
			"/v1/chat/completions",
			"deepseek-v4-flash",
			"https://opencode.ai/zen/go/v1/chat/completions",
		},
		{
			// 规范化后：/v1 结尾基础 URL + responses 请求
			"v1 base url + responses request",
			"https://opencode.ai/zen/go/v1",
			"/v1/responses",
			"deepseek-v4-flash",
			"https://opencode.ai/zen/go/v1/responses",
		},
		{
			// {model} 占位符替换
			"model placeholder",
			"https://example.com/zen/{model}/v1/chat/completions",
			"/v1/chat/completions",
			"mimo-v2.5",
			"https://example.com/zen/mimo-v2.5/v1/chat/completions",
		},
		{
			// 未知端点类型（自定义路径）→ 保持原样
			"custom path + unknown endpoint",
			"https://example.com/custom/endpoint",
			"/v1/custom/foo",
			"deepseek-v4-flash",
			"https://example.com/custom/endpoint",
		},
		{
			// 完整 responses URL + responses 请求
			"full responses url + responses request",
			"https://opencode.ai/zen/go/v1/responses",
			"/v1/responses",
			"deepseek-v4-flash",
			"https://opencode.ai/zen/go/v1/responses",
		},
		{
			// 完整 responses URL + chat 请求 → 自动换 /v1/chat/completions
			"full responses url + chat request",
			"https://opencode.ai/zen/go/v1/responses",
			"/v1/chat/completions",
			"deepseek-v4-flash",
			"https://opencode.ai/zen/go/v1/chat/completions",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ResolveCustomChannelURL(tt.baseURL, tt.requestPath, tt.model))
		})
	}
}
