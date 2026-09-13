package model

import (
	"os"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestValidateModelCostJSONRejectsUnsafeValues(t *testing.T) {
	require.Error(t, ValidateModelCostJSON(`{"gpt-4o":{"enabled":true,"input_per_1m":-1}}`))
	require.Error(t, ValidateModelCostJSON(`{"gpt-4o":{"enabled":true,"output_per_1m":1e20}}`))
	require.Error(t, ValidateModelCostJSON(`{"gpt-4o":{"enabled":true,"image_token_per_1m":1e20}}`))
	require.Error(t, ValidateModelCostJSON(`{"gpt-4o":{"enabled":true,"currency":"US DOLLAR"}}`))
	require.NoError(t, ValidateModelCostJSON(`{"gpt-4o":{"enabled":true,"input_per_1m":2.5,"output_per_1m":10}}`))
}

func TestCalculateCostAccountingIgnoresRuntimeInvalidCostEntry(t *testing.T) {
	previous := common.OptionMap
	common.OptionMapRWMutex.Lock()
	common.OptionMap = map[string]string{
		"ModelCost": `{"bad":{"enabled":true,"input_per_1m":1}}`,
	}
	common.OptionMapRWMutex.Unlock()
	t.Cleanup(func() {
		common.OptionMapRWMutex.Lock()
		common.OptionMap = previous
		common.OptionMapRWMutex.Unlock()
	})

	// A malformed in-memory entry can only be exercised by replacing the
	// parsed option after loading; the public JSON validator rejects it first.
	common.OptionMapRWMutex.Lock()
	common.OptionMap["ModelCost"] = `{"bad":{"enabled":true,"currency":"TOO-LONG-CURRENCY"}}`
	common.OptionMapRWMutex.Unlock()
	accounting := CalculateCostAccounting("bad", 0, 100, true, 1, 0, 0, 0, 0, 0, 0, 0, 0)
	assert.Empty(t, accounting.Source)
}

func TestCalculateCostAccountingUsesChannelPrecedenceAndImmutableRevenue(t *testing.T) {
	previous := common.OptionMap
	common.OptionMapRWMutex.Lock()
	common.OptionMap = map[string]string{
		"ModelCost": `{"default":{"enabled":true,"input_per_1m":1},"gpt-4o":{"enabled":true,"input_per_1m":2,"output_per_1m":4},"channel:7:gpt-4o":{"enabled":true,"input_per_1m":3,"output_per_1m":6}}`,
	}
	common.OptionMapRWMutex.Unlock()
	t.Cleanup(func() {
		common.OptionMapRWMutex.Lock()
		common.OptionMap = previous
		common.OptionMapRWMutex.Unlock()
	})

	oldQuotaPerUnit := common.QuotaPerUnit
	common.QuotaPerUnit = 500000
	t.Cleanup(func() { common.QuotaPerUnit = oldQuotaPerUnit })

	accounting := CalculateCostAccounting("gpt-4o", 7, 500000, true, 1000, 500, 0, 0, 0, 0, 0, 0, 0)
	assert.Equal(t, int64(6000), accounting.ActualCostMicros)
	assert.Equal(t, int64(1000000), accounting.RevenueMicros)
	assert.Equal(t, "channel:7:gpt-4o", accounting.Source)

	withoutUsage := CalculateCostAccounting("gpt-4o", 7, 500000, false, 1000, 500, 0, 0, 0, 0, 0, 0, 0)
	assert.Zero(t, withoutUsage.ActualCostMicros)
	assert.Equal(t, int64(1000000), withoutUsage.RevenueMicros)
}

func TestGetCostStatisticsAggregatesRequestsAndExcludesFailuresFromRevenue(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&Log{}))
	previousLogDB := LOG_DB
	previousMainType, previousLogType := common.MainDatabaseType(), common.LogDatabaseType()
	LOG_DB = db
	common.SetDatabaseTypes(common.DatabaseTypeSQLite, common.DatabaseTypeSQLite)
	initCol()
	t.Cleanup(func() {
		LOG_DB = previousLogDB
		common.SetDatabaseTypes(previousMainType, previousLogType)
		initCol()
	})

	require.NoError(t, db.Create(&Log{ModelName: "gpt-4o", ChannelId: 7, Type: LogTypeConsume, CreatedAt: 100, PromptTokens: 100, CompletionTokens: 20, Quota: 500000, ActualCostMicros: 3000, RevenueMicros: 1000000, CostCurrency: "USD", BillingEvent: "request", UsageAvailable: true}).Error)
	require.NoError(t, db.Create(&Log{ModelName: "gpt-4o", ChannelId: 7, Type: LogTypeError, CreatedAt: 101, Content: "upstream failed"}).Error)
	require.NoError(t, db.Create(&Log{ModelName: "gpt-4o", ChannelId: 7, Type: LogTypeRefund, CreatedAt: 102, Quota: 500000, RevenueMicros: -1000000, CostCurrency: "USD", BillingEvent: "refund"}).Error)

	stats, err := GetCostStatistics(CostStatisticsFilter{StartTimestamp: 1, EndTimestamp: 200, ModelName: "gpt-4o"})
	require.NoError(t, err)
	require.Len(t, stats.Rows, 1)
	assert.Equal(t, int64(1), stats.RequestCount)
	assert.Equal(t, int64(1), stats.SuccessCount)
	assert.Equal(t, int64(1), stats.FailureCount)
	assert.Equal(t, int64(120), stats.TotalTokens)
	assert.Equal(t, int64(3000), stats.ActualCostMicros)
	assert.Equal(t, int64(0), stats.RevenueMicros)
	assert.Equal(t, int64(-3000), stats.GrossProfitMicros)
}

func testModelCostLogMigration(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.AutoMigrate(&Log{}))
	require.NoError(t, db.AutoMigrate(&Log{}))
	row := &Log{ModelName: "migration-cost-model", CreatedAt: 1, Type: LogTypeConsume, BillingEvent: "request", ActualCostMicros: 123, RevenueMicros: 456}
	require.NoError(t, db.Create(row).Error)
	var found Log
	require.NoError(t, db.First(&found, row.Id).Error)
	assert.Equal(t, int64(123), found.ActualCostMicros)
	assert.Equal(t, int64(456), found.RevenueMicros)
	require.NoError(t, db.Delete(&Log{}, row.Id).Error)
}

func TestModelCostLogMigrationMySQL(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("TEST_MYSQL_DSN"))
	if dsn == "" {
		t.Skip("TEST_MYSQL_DSN is not configured")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	testModelCostLogMigration(t, db)
}

func TestModelCostLogMigrationPostgreSQL(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("TEST_POSTGRES_DSN"))
	if dsn == "" {
		t.Skip("TEST_POSTGRES_DSN is not configured")
	}
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: dsn, PreferSimpleProtocol: true}), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	testModelCostLogMigration(t, db)
}

func TestCalculateCostAccountingSeparatesModalityTokens(t *testing.T) {
	previous := common.OptionMap
	common.OptionMapRWMutex.Lock()
	common.OptionMap = map[string]string{
		"ModelCost": `{"modal-model":{"enabled":true,"input_per_1m":1,"output_per_1m":2,"audio_input_per_1m":10,"audio_output_per_1m":20,"image_token_per_1m":30}}`,
	}
	common.OptionMapRWMutex.Unlock()
	t.Cleanup(func() {
		common.OptionMapRWMutex.Lock()
		common.OptionMap = previous
		common.OptionMapRWMutex.Unlock()
	})

	accounting := CalculateCostAccountingWithUsage("modal-model", 0, 0, CostUsage{
		UsageAvailable:    true,
		PromptTokens:      100,
		CompletionTokens:  50,
		AudioInputTokens:  10,
		AudioOutputTokens: 5,
		ImageTokens:       20,
	})
	// Generic text: (100 - 10 - 20) * 1 + (50 - 5) * 2;
	// modality tokens: 10 * 10 + 5 * 20 + 20 * 30 micro-USD.
	assert.Equal(t, int64(960), accounting.ActualCostMicros)
}

func TestCalculateCostAccountingUsesVideoResolutionCost(t *testing.T) {
	common.OptionMapRWMutex.Lock()
	if common.OptionMap == nil {
		common.OptionMap = make(map[string]string)
	}
	saved, hadSaved := common.OptionMap["ModelCost"]
	common.OptionMapRWMutex.Unlock()
	t.Cleanup(func() {
		common.OptionMapRWMutex.Lock()
		defer common.OptionMapRWMutex.Unlock()
		if hadSaved {
			common.OptionMap["ModelCost"] = saved
		} else {
			delete(common.OptionMap, "ModelCost")
		}
	})
	common.OptionMapRWMutex.Lock()
	common.OptionMap["ModelCost"] = `{"video-model":{"enabled":true,"video_per_second":0.01,"video_per_second_by_resolution":{"768p":0.15,"1080p":0.2,"1440p":0.3,"4k":0.35}}}`
	common.OptionMapRWMutex.Unlock()

	accounting := CalculateCostAccountingWithUsage("video-model", 0, 0, CostUsage{
		UsageAvailable:  true,
		VideoSeconds:    2,
		VideoResolution: "2K",
	})

	if accounting.ActualCostMicros != 600000 {
		t.Fatalf("resolution cost = %d, want %d", accounting.ActualCostMicros, int64(600000))
	}
}

func TestValidateModelCostJSONRejectsInvalidVideoResolutionCost(t *testing.T) {
	err := ValidateModelCostJSON(`{"video-model":{"enabled":true,"video_per_second_by_resolution":{"1080p":-1}}}`)
	if err == nil {
		t.Fatal("expected invalid resolution cost to be rejected")
	}
}
