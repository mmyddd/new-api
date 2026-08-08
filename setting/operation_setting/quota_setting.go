package operation_setting

import "github.com/QuantumNous/new-api/setting/config"

type QuotaSetting struct {
	EnableFreeModelPreConsume bool `json:"enable_free_model_pre_consume"` // 是否对免费模型启用预消耗
	// SkipClientGoneBilling 不统计断流数据：流式请求客户端断开（client_gone）
	// 时不再计费、不写入消费日志与统计
	SkipClientGoneBilling bool `json:"skip_client_gone_billing"`
}

// 默认配置
var quotaSetting = QuotaSetting{
	EnableFreeModelPreConsume: true,
	SkipClientGoneBilling:     false,
}

func init() {
	// 注册到全局配置管理器
	config.GlobalConfig.Register("quota_setting", &quotaSetting)
}

func GetQuotaSetting() *QuotaSetting {
	return &quotaSetting
}
