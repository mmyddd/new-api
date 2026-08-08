package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func useQuotaDataTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	previousDB := DB
	previousType := common.MainDatabaseType()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&QuotaData{}))
	DB = db
	common.SetMainDatabaseType(common.DatabaseTypeSQLite)
	t.Cleanup(func() {
		DB = previousDB
		common.SetMainDatabaseType(previousType)
	})
	return db
}

// TestLogQuotaDataAccumulatesTokenBreakdown 保护数据看板 token 构成（输入/缓存读取/输出）
// 与总 Token 数同源聚合：多条日志先内存累加、落盘后再次写入走增量更新，
// 且 token_used 恒等于 input+output。
func TestLogQuotaDataAccumulatesTokenBreakdown(t *testing.T) {
	db := useQuotaDataTestDB(t)
	CacheQuotaDataLock.Lock()
	CacheQuotaData = make(map[string]*QuotaData)
	CacheQuotaDataLock.Unlock()

	base := QuotaDataLogParams{
		UserID:    1,
		Username:  "qatest",
		ModelName: "gpt-test",
		Quota:     1000,
		CreatedAt: 1_700_000_000,
		UseGroup:  "default",
		TokenID:   1,
		ChannelID: 1,
		NodeName:  "node",
	}

	first := base
	first.TokenUsed = 100
	first.InputTokens = 80
	first.CacheTokens = 60
	first.OutputTokens = 20
	LogQuotaData(first)

	second := base
	second.TokenUsed = 200
	second.InputTokens = 150
	second.CacheTokens = 120
	second.OutputTokens = 50
	LogQuotaData(second)

	SaveQuotaDataCache()

	var row QuotaData
	require.NoError(t, db.Table("quota_data").Where("username = ?", "qatest").First(&row).Error)
	assert.Equal(t, 2, row.Count)
	assert.Equal(t, 300, row.TokenUsed)
	assert.Equal(t, 230, row.InputTokens)
	assert.Equal(t, 180, row.CacheTokens)
	assert.Equal(t, 70, row.OutputTokens)
	assert.Equal(t, row.TokenUsed, row.InputTokens+row.OutputTokens)

	// 同桶再次写入走 increaseQuotaData 增量路径
	third := base
	third.TokenUsed = 50
	third.InputTokens = 40
	third.CacheTokens = 30
	third.OutputTokens = 10
	LogQuotaData(third)
	SaveQuotaDataCache()

	var row2 QuotaData
	require.NoError(t, db.Table("quota_data").Where("username = ?", "qatest").First(&row2).Error)
	assert.Equal(t, 3, row2.Count)
	assert.Equal(t, 350, row2.TokenUsed)
	assert.Equal(t, 270, row2.InputTokens)
	assert.Equal(t, 210, row2.CacheTokens)
	assert.Equal(t, 80, row2.OutputTokens)
	assert.Equal(t, row2.TokenUsed, row2.InputTokens+row2.OutputTokens)
}
