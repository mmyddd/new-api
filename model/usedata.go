package model

import (
	"fmt"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

// QuotaData 柱状图数据
type QuotaData struct {
	Id        int    `json:"id"`
	UserID    int    `json:"user_id" gorm:"index"`
	Username  string `json:"username" gorm:"index:idx_qdt_model_user_name,priority:2;size:64;default:''"`
	ModelName string `json:"model_name" gorm:"index:idx_qdt_model_user_name,priority:1;size:64;default:''"`
	CreatedAt int64  `json:"created_at" gorm:"bigint;index:idx_qdt_created_at,priority:2"`
	UseGroup  string `json:"use_group" gorm:"index;size:64;default:''"`
	TokenID   int    `json:"token_id" gorm:"index;default:0"`
	ChannelID int    `json:"channel_id" gorm:"index;default:0"`
	NodeName  string `json:"node_name" gorm:"index;size:64;default:''"`
	TokenUsed int    `json:"token_used" gorm:"default:0"`
	// 时间范围内的 token 构成（输入/缓存读取/输出），供数据看板卡片小字与总 Token 数同源展示
	InputTokens  int `json:"input_tokens" gorm:"default:0"`
	CacheTokens  int `json:"cache_tokens" gorm:"default:0"`
	OutputTokens int `json:"output_tokens" gorm:"default:0"`
	Count        int `json:"count" gorm:"default:0"`
	Quota        int `json:"quota" gorm:"default:0"`
}

type QuotaDataLogParams struct {
	UserID    int
	Username  string
	ModelName string
	Quota     int
	CreatedAt int64
	TokenUsed int
	// token 构成：InputTokens 为总输入（含缓存），CacheTokens 为缓存读取
	InputTokens  int
	CacheTokens  int
	OutputTokens int
	UseGroup     string
	TokenID      int
	ChannelID    int
	NodeName     string
}

func UpdateQuotaData() {
	for {
		if common.DataExportEnabled {
			common.SysLog("正在更新数据看板数据...")
			SaveQuotaDataCache()
		}
		time.Sleep(time.Duration(common.DataExportInterval) * time.Minute)
	}
}

var CacheQuotaData = make(map[string]*QuotaData)
var CacheQuotaDataLock = sync.Mutex{}

func logQuotaDataCache(quotaData *QuotaData) {
	key := fmt.Sprintf("%d\x00%s\x00%s\x00%d\x00%s\x00%d\x00%d\x00%s",
		quotaData.UserID,
		quotaData.Username,
		quotaData.ModelName,
		quotaData.CreatedAt,
		quotaData.UseGroup,
		quotaData.TokenID,
		quotaData.ChannelID,
		quotaData.NodeName,
	)
	count := quotaData.Count
	quota := quotaData.Quota
	tokenUsed := quotaData.TokenUsed
	inputTokens := quotaData.InputTokens
	cacheTokens := quotaData.CacheTokens
	outputTokens := quotaData.OutputTokens
	cachedQuotaData, ok := CacheQuotaData[key]
	if ok {
		cachedQuotaData.Count += count
		cachedQuotaData.Quota += quota
		cachedQuotaData.TokenUsed += tokenUsed
		cachedQuotaData.InputTokens += inputTokens
		cachedQuotaData.CacheTokens += cacheTokens
		cachedQuotaData.OutputTokens += outputTokens
		quotaData = cachedQuotaData
	}
	CacheQuotaData[key] = quotaData
}

func LogQuotaData(params QuotaDataLogParams) {
	// 只精确到小时
	createdAt := params.CreatedAt - (params.CreatedAt % 3600)
	quotaData := &QuotaData{
		UserID:       params.UserID,
		Username:     params.Username,
		ModelName:    params.ModelName,
		CreatedAt:    createdAt,
		UseGroup:     params.UseGroup,
		TokenID:      params.TokenID,
		ChannelID:    params.ChannelID,
		NodeName:     params.NodeName,
		Count:        1,
		Quota:        params.Quota,
		TokenUsed:    params.TokenUsed,
		InputTokens:  params.InputTokens,
		CacheTokens:  params.CacheTokens,
		OutputTokens: params.OutputTokens,
	}

	CacheQuotaDataLock.Lock()
	defer CacheQuotaDataLock.Unlock()
	logQuotaDataCache(quotaData)
}

func SaveQuotaDataCache() {
	CacheQuotaDataLock.Lock()
	defer CacheQuotaDataLock.Unlock()
	size := len(CacheQuotaData)
	// 如果缓存中有数据，就保存到数据库中
	// 1. 先查询数据库中是否有数据
	// 2. 如果有数据，就更新数据
	// 3. 如果没有数据，就插入数据
	for _, quotaData := range CacheQuotaData {
		quotaDataDB := &QuotaData{}
		DB.Table("quota_data").
			Where("user_id = ? and username = ? and model_name = ? and created_at = ? and use_group = ? and token_id = ? and channel_id = ? and node_name = ?",
				quotaData.UserID, quotaData.Username, quotaData.ModelName, quotaData.CreatedAt, quotaData.UseGroup, quotaData.TokenID, quotaData.ChannelID, quotaData.NodeName).
			First(quotaDataDB)
		if quotaDataDB.Id > 0 {
			//quotaDataDB.Count += quotaData.Count
			//quotaDataDB.Quota += quotaData.Quota
			//DB.Table("quota_data").Save(quotaDataDB)
			increaseQuotaData(quotaData)
		} else {
			DB.Table("quota_data").Create(quotaData)
		}
	}
	CacheQuotaData = make(map[string]*QuotaData)
	common.SysLog(fmt.Sprintf("保存数据看板数据成功，共保存%d条数据", size))
}

func increaseQuotaData(quotaData *QuotaData) {
	err := DB.Table("quota_data").
		Where("user_id = ? and username = ? and model_name = ? and created_at = ? and use_group = ? and token_id = ? and channel_id = ? and node_name = ?",
			quotaData.UserID, quotaData.Username, quotaData.ModelName, quotaData.CreatedAt, quotaData.UseGroup, quotaData.TokenID, quotaData.ChannelID, quotaData.NodeName).
		Updates(map[string]interface{}{
			"count":         gorm.Expr("count + ?", quotaData.Count),
			"quota":         gorm.Expr("quota + ?", quotaData.Quota),
			"token_used":    gorm.Expr("token_used + ?", quotaData.TokenUsed),
			"input_tokens":  gorm.Expr("input_tokens + ?", quotaData.InputTokens),
			"cache_tokens":  gorm.Expr("cache_tokens + ?", quotaData.CacheTokens),
			"output_tokens": gorm.Expr("output_tokens + ?", quotaData.OutputTokens),
		}).Error
	if err != nil {
		common.SysLog(fmt.Sprintf("increaseQuotaData error: %s", err))
	}
}

// resolveTokenIDsByName returns the ids of every token whose name matches
// exactly. When userID is positive the lookup is scoped to that user's own
// tokens, so self-service endpoints can never leak another user's data.
// An empty slice means no token has that name.
func resolveTokenIDsByName(name string, userID int) ([]int, error) {
	var ids []int
	query := DB.Model(&Token{}).Where("name = ?", name)
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if err := query.Pluck("id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

func GetQuotaDataByUsername(username string, startTime int64, endTime int64, tokenName string) (quotaData []*QuotaData, err error) {
	var quotaDatas []*QuotaData
	// 从quota_data表中查询数据
	query := DB.Table("quota_data").
		Select("user_id, username, model_name, created_at, sum(count) as count, sum(quota) as quota, sum(token_used) as token_used, sum(input_tokens) as input_tokens, sum(cache_tokens) as cache_tokens, sum(output_tokens) as output_tokens").
		Where("username = ? and created_at >= ? and created_at <= ?", username, startTime, endTime)
	if tokenName != "" {
		tokenIDs, err := resolveTokenIDsByName(tokenName, 0)
		if err != nil {
			return nil, err
		}
		if len(tokenIDs) == 0 {
			return quotaDatas, nil
		}
		query = query.Where("token_id IN ?", tokenIDs)
	}
	err = query.Group("user_id, username, model_name, created_at").
		Find(&quotaDatas).Error
	return quotaDatas, err
}

func GetQuotaDataByUserId(userId int, startTime int64, endTime int64, tokenName string) (quotaData []*QuotaData, err error) {
	var quotaDatas []*QuotaData
	// 从quota_data表中查询数据
	query := DB.Table("quota_data").
		Select("user_id, username, model_name, created_at, sum(count) as count, sum(quota) as quota, sum(token_used) as token_used, sum(input_tokens) as input_tokens, sum(cache_tokens) as cache_tokens, sum(output_tokens) as output_tokens").
		Where("user_id = ? and created_at >= ? and created_at <= ?", userId, startTime, endTime)
	if tokenName != "" {
		// Self-service filter is scoped to the user's own tokens so a token
		// name owned by someone else cannot leak their usage data.
		tokenIDs, err := resolveTokenIDsByName(tokenName, userId)
		if err != nil {
			return nil, err
		}
		if len(tokenIDs) == 0 {
			return quotaDatas, nil
		}
		query = query.Where("token_id IN ?", tokenIDs)
	}
	err = query.Group("user_id, username, model_name, created_at").
		Find(&quotaDatas).Error
	return quotaDatas, err
}

func GetQuotaDataGroupByUser(startTime int64, endTime int64) (quotaData []*QuotaData, err error) {
	var quotaDatas []*QuotaData
	err = DB.Table("quota_data").
		Select("username, created_at, sum(count) as count, sum(quota) as quota, sum(token_used) as token_used, sum(input_tokens) as input_tokens, sum(cache_tokens) as cache_tokens, sum(output_tokens) as output_tokens").
		Where("created_at >= ? and created_at <= ?", startTime, endTime).
		Group("username, created_at").
		Find(&quotaDatas).Error
	return quotaDatas, err
}

func GetAllQuotaDates(startTime int64, endTime int64, username string, tokenName string) (quotaData []*QuotaData, err error) {
	if username != "" {
		return GetQuotaDataByUsername(username, startTime, endTime, tokenName)
	}
	var quotaDatas []*QuotaData
	// 从quota_data表中查询数据
	// only select model_name, sum(count) as count, sum(quota) as quota, model_name, created_at from quota_data group by model_name, created_at;
	query := DB.Table("quota_data").Select("model_name, sum(count) as count, sum(quota) as quota, sum(token_used) as token_used, sum(input_tokens) as input_tokens, sum(cache_tokens) as cache_tokens, sum(output_tokens) as output_tokens, created_at").Where("created_at >= ? and created_at <= ?", startTime, endTime)
	if tokenName != "" {
		tokenIDs, err := resolveTokenIDsByName(tokenName, 0)
		if err != nil {
			return nil, err
		}
		if len(tokenIDs) == 0 {
			return quotaDatas, nil
		}
		query = query.Where("token_id IN ?", tokenIDs)
	}
	err = query.Group("model_name, created_at").Find(&quotaDatas).Error
	return quotaDatas, err
}
