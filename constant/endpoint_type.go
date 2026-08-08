package constant

import "github.com/QuantumNous/new-api/relaykit/types"

// EndpointType moved to types with the conversion kit; aliases keep host
// code compiling unchanged.
type EndpointType = types.EndpointType

const (
	EndpointTypeOpenAI                = types.EndpointTypeOpenAI
	EndpointTypeOpenAIResponse        = types.EndpointTypeOpenAIResponse
	EndpointTypeOpenAIResponseCompact = types.EndpointTypeOpenAIResponseCompact
	EndpointTypeOpenAIAlphaSearch     = types.EndpointTypeOpenAIAlphaSearch
	EndpointTypeAnthropic             = types.EndpointTypeAnthropic
	EndpointTypeGemini                = types.EndpointTypeGemini
	EndpointTypeJinaRerank            = types.EndpointTypeJinaRerank
	EndpointTypeImageGeneration       = types.EndpointTypeImageGeneration
	EndpointTypeEmbeddings            = types.EndpointTypeEmbeddings
	EndpointTypeOpenAIVideo           = types.EndpointTypeOpenAIVideo
)

// AllEndpointTypes 列出全部已知端点类型，用于渠道保存时校验声明的端点类型。
var AllEndpointTypes = []EndpointType{
	EndpointTypeOpenAI,
	EndpointTypeOpenAIResponse,
	EndpointTypeOpenAIResponseCompact,
	EndpointTypeOpenAIAlphaSearch,
	EndpointTypeAnthropic,
	EndpointTypeGemini,
	EndpointTypeJinaRerank,
	EndpointTypeImageGeneration,
	EndpointTypeEmbeddings,
	EndpointTypeOpenAIVideo,
}
