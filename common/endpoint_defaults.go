package common

import (
	"strings"

	"github.com/QuantumNous/new-api/constant"
)

// EndpointInfo 描述单个端点的默认请求信息
// path: 上游路径
// method: HTTP 请求方式，例如 POST/GET
// 目前均为 POST，后续可扩展
//
// json 标签用于直接序列化到 API 输出
// 例如：{"path":"/v1/chat/completions","method":"POST"}

type EndpointInfo struct {
	Path   string `json:"path"`
	Method string `json:"method"`
}

// defaultEndpointInfoMap 保存内置端点的默认 Path 与 Method
var defaultEndpointInfoMap = map[constant.EndpointType]EndpointInfo{
	constant.EndpointTypeOpenAI:                {Path: "/v1/chat/completions", Method: "POST"},
	constant.EndpointTypeOpenAIResponse:        {Path: "/v1/responses", Method: "POST"},
	constant.EndpointTypeOpenAIResponseCompact: {Path: "/v1/responses/compact", Method: "POST"},
	constant.EndpointTypeOpenAIAlphaSearch:     {Path: "/v1/alpha/search", Method: "POST"},
	constant.EndpointTypeAnthropic:             {Path: "/v1/messages", Method: "POST"},
	constant.EndpointTypeGemini:                {Path: "/v1beta/models/{model}:generateContent", Method: "POST"},
	constant.EndpointTypeJinaRerank:            {Path: "/v1/rerank", Method: "POST"},
	constant.EndpointTypeImageGeneration:       {Path: "/v1/images/generations", Method: "POST"},
	constant.EndpointTypeEmbeddings:            {Path: "/v1/embeddings", Method: "POST"},
}

// GetDefaultEndpointInfo 返回指定端点类型的默认信息以及是否存在
func GetDefaultEndpointInfo(et constant.EndpointType) (EndpointInfo, bool) {
	info, ok := defaultEndpointInfoMap[et]
	return info, ok
}

// Path2EndpointType 将请求路径映射为对应的端点类型。
// 无法识别的路径（音频、midjourney、/v1/models 等）返回空值，表示不参与端点类型过滤。
func Path2EndpointType(path string) constant.EndpointType {
	switch {
	case strings.HasPrefix(path, "/v1/chat/completions"):
		return constant.EndpointTypeOpenAI
	case strings.HasPrefix(path, "/v1/responses/compact"):
		return constant.EndpointTypeOpenAIResponseCompact
	case strings.HasPrefix(path, "/v1/responses"):
		return constant.EndpointTypeOpenAIResponse
	case strings.HasPrefix(path, "/v1/alpha/search"):
		return constant.EndpointTypeOpenAIAlphaSearch
	case strings.HasPrefix(path, "/v1/messages"):
		return constant.EndpointTypeAnthropic
	case strings.HasPrefix(path, "/v1/rerank"):
		return constant.EndpointTypeJinaRerank
	case strings.HasPrefix(path, "/v1/images/generations"):
		return constant.EndpointTypeImageGeneration
	case strings.HasPrefix(path, "/v1/embeddings"):
		return constant.EndpointTypeEmbeddings
	case strings.HasPrefix(path, "/v1/video/generations"):
		return constant.EndpointTypeOpenAIVideo
	default:
		if strings.HasPrefix(path, "/v1beta/models/") &&
			(strings.Contains(path, ":generateContent") || strings.Contains(path, ":streamGenerateContent")) {
			return constant.EndpointTypeGemini
		}
		return ""
	}
}

// NormalizeBaseURL 去除基础 URL 末尾的 /v1（含 /v1/ 变体）。
// 上游请求路径本身以 /v1 开头，由各适配器拼接在基础 URL 之后；基础 URL 保留 /v1
// 会产生 /v1/v1/... 的重复路径，因此拼接前统一去除。存储值保持用户录入不变。
func NormalizeBaseURL(baseURL string) string {
	trimmed := strings.TrimRight(baseURL, "/")
	if strings.HasSuffix(trimmed, "/v1") {
		return strings.TrimSuffix(trimmed, "/v1")
	}
	return baseURL
}
