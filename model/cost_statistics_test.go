package model

import (
	"fmt"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupCostStatisticsModelTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	previousLogDB := LOG_DB
	previousMainType := common.MainDatabaseType()
	previousLogType := common.LogDatabaseType()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&Log{}))

	LOG_DB = db
	common.SetDatabaseTypes(common.DatabaseTypeSQLite, common.DatabaseTypeSQLite)
	initCol()
	t.Cleanup(func() {
		LOG_DB = previousLogDB
		common.SetDatabaseTypes(previousMainType, previousLogType)
		initCol()
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	return db
}

func TestGetCostStatisticsAppliesAllFilters(t *testing.T) {
	db := setupCostStatisticsModelTestDB(t)

	logs := []*Log{
		{ModelName: "gpt-filter", Username: "alice", ChannelId: 7, Group: "vip", CreatedAt: 100, Type: LogTypeConsume, BillingEvent: "request", RevenueMicros: 1000},
		{ModelName: "gpt-filter", Username: "bob", ChannelId: 7, Group: "vip", CreatedAt: 100, Type: LogTypeConsume, BillingEvent: "request", RevenueMicros: 2000},
		{ModelName: "other-model", Username: "alice", ChannelId: 7, Group: "vip", CreatedAt: 100, Type: LogTypeConsume, BillingEvent: "request", RevenueMicros: 3000},
		{ModelName: "gpt-filter", Username: "alice", ChannelId: 8, Group: "vip", CreatedAt: 100, Type: LogTypeConsume, BillingEvent: "request", RevenueMicros: 4000},
		{ModelName: "gpt-filter", Username: "alice", ChannelId: 7, Group: "default", CreatedAt: 100, Type: LogTypeConsume, BillingEvent: "request", RevenueMicros: 5000},
		{ModelName: "gpt-filter", Username: "alice", ChannelId: 7, Group: "vip", CreatedAt: 99, Type: LogTypeConsume, BillingEvent: "request", RevenueMicros: 6000},
		{ModelName: "gpt-filter", Username: "alice", ChannelId: 7, Group: "vip", CreatedAt: 200, Type: LogTypeConsume, BillingEvent: "request", RevenueMicros: 7000},
	}
	for _, log := range logs {
		require.NoError(t, db.Create(log).Error)
	}

	stats, err := GetCostStatistics(CostStatisticsFilter{
		StartTimestamp: 100,
		EndTimestamp:   199,
		ModelName:      "gpt-filter",
		Username:       "alice",
		ChannelID:      7,
		Group:          "vip",
	})
	require.NoError(t, err)
	require.Len(t, stats.Rows, 1)
	assert.Equal(t, int64(1), stats.Rows[0].RequestCount)
	assert.Equal(t, int64(1000), stats.Rows[0].RevenueMicros)
}

func TestGetCostStatisticsExcludesFailureRevenueButIncludesRefunds(t *testing.T) {
	db := setupCostStatisticsModelTestDB(t)

	require.NoError(t, db.Create(&Log{
		ModelName: "billing-model", ChannelId: 3, CreatedAt: 100, Type: LogTypeConsume,
		BillingEvent: "request", ActualCostMicros: 300, RevenueMicros: 1000,
	}).Error)
	require.NoError(t, db.Create(&Log{
		ModelName: "billing-model", ChannelId: 3, CreatedAt: 101, Type: LogTypeError,
		BillingEvent: "request", ActualCostMicros: 900, RevenueMicros: 500,
	}).Error)
	require.NoError(t, db.Create(&Log{
		ModelName: "billing-model", ChannelId: 3, CreatedAt: 102, Type: LogTypeRefund,
		BillingEvent: "refund", ActualCostMicros: 700, RevenueMicros: -250,
	}).Error)

	stats, err := GetCostStatistics(CostStatisticsFilter{StartTimestamp: 100, EndTimestamp: 102})
	require.NoError(t, err)
	require.Len(t, stats.Rows, 1)
	row := stats.Rows[0]
	assert.Equal(t, int64(1), row.RequestCount)
	assert.Equal(t, int64(1), row.SuccessCount)
	assert.Equal(t, int64(1), row.FailureCount)
	assert.Equal(t, int64(750), row.RevenueMicros, "failed requests must not create revenue; refunds must reduce settled revenue")
	assert.Equal(t, int64(300), row.ActualCostMicros, "refund and failure rows must not add upstream request cost")
	assert.Equal(t, int64(450), row.GrossProfitMicros)
}

func TestGetCostStatisticsDoesNotMixCurrenciesInSummaryAmounts(t *testing.T) {
	db := setupCostStatisticsModelTestDB(t)

	require.NoError(t, db.Create(&Log{
		ModelName: "currency-model", ChannelId: 4, CreatedAt: 100, Type: LogTypeConsume,
		BillingEvent: "request", CostCurrency: "USD", ActualCostMicros: 400, RevenueMicros: 1000,
	}).Error)
	require.NoError(t, db.Create(&Log{
		ModelName: "currency-model", ChannelId: 4, CreatedAt: 101, Type: LogTypeConsume,
		BillingEvent: "request", CostCurrency: "EUR", ActualCostMicros: 800, RevenueMicros: 2000,
	}).Error)

	stats, err := GetCostStatistics(CostStatisticsFilter{StartTimestamp: 100, EndTimestamp: 101})
	require.NoError(t, err)
	require.Len(t, stats.Rows, 2)
	assert.True(t, stats.MixedCurrency)
	assert.Empty(t, stats.Currency)
	assert.Zero(t, stats.ActualCostMicros, "summary amounts must not be added across currencies")
	assert.Zero(t, stats.RevenueMicros, "summary amounts must not be added across currencies")
	assert.Zero(t, stats.GrossProfitMicros, "summary amounts must not be added across currencies")

	for _, row := range stats.Rows {
		switch row.Currency {
		case "USD":
			assert.Equal(t, int64(600), row.GrossProfitMicros)
		case "EUR":
			assert.Equal(t, int64(1200), row.GrossProfitMicros)
		default:
			t.Fatalf("unexpected currency %q", row.Currency)
		}
	}
}
