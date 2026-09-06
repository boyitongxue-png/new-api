package model

import (
	"fmt"
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

func setupCostStatisticsDetailsDB(t *testing.T) *gorm.DB {
	t.Helper()
	previousLogDB := LOG_DB
	previousMainType, previousLogType := common.MainDatabaseType(), common.LogDatabaseType()
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

func TestGetCostStatisticsDetailsFiltersAndPaginatesAccountingRows(t *testing.T) {
	db := setupCostStatisticsDetailsDB(t)
	require.NoError(t, db.Create(&Log{
		Id:                1,
		UserId:            42,
		CreatedAt:         120,
		Type:              LogTypeConsume,
		BillingEvent:      "request",
		ModelName:         "gpt-4o",
		Username:          "alice",
		ChannelId:         7,
		Group:             "vip",
		RequestId:         "req-2",
		UpstreamRequestId: "up-2",
		Quota:             200,
		PromptTokens:      100,
		CompletionTokens:  20,
		ActualCostMicros:  300,
		RevenueMicros:     800,
		CostCurrency:      "USD",
		UsageAvailable:    true,
		Content:           "do not expose this prompt",
		Other:             `{"secret":"do not expose"}`,
		Ip:                "192.0.2.10",
		TokenId:           99,
		TokenName:         "secret-token",
	}).Error)
	require.NoError(t, db.Create(&Log{
		Id:            2,
		CreatedAt:     110,
		Type:          LogTypeRefund,
		BillingEvent:  "refund",
		ModelName:     "gpt-4o",
		Username:      "alice",
		ChannelId:     7,
		Group:         "vip",
		RequestId:     "req-1",
		Quota:         -200,
		RevenueMicros: -800,
		CostCurrency:  "USD",
	}).Error)
	// This row proves all dimensions are applied, including currency and request ID.
	require.NoError(t, db.Create(&Log{
		Id:           3,
		CreatedAt:    130,
		Type:         LogTypeConsume,
		BillingEvent: "request",
		ModelName:    "gpt-4o",
		Username:     "bob",
		ChannelId:    8,
		Group:        "default",
		RequestId:    "req-other",
		CostCurrency: "EUR",
	}).Error)
	require.NoError(t, db.Create(&Log{
		Id:           4,
		CreatedAt:    140,
		Type:         LogTypeManage,
		ModelName:    "gpt-4o",
		Username:     "alice",
		ChannelId:    7,
		Group:        "vip",
		RequestId:    "req-2",
		CostCurrency: "USD",
	}).Error)

	filter := CostStatisticsDetailFilter{
		CostStatisticsFilter: CostStatisticsFilter{
			StartTimestamp: 100,
			EndTimestamp:   125,
			ModelName:      "gpt-4o",
			Username:       "alice",
			ChannelID:      7,
			Group:          "vip",
		},
		Currency:  "USD",
		RequestID: "req-2",
	}
	rows, total, err := GetCostStatisticsDetails(filter, 0, 1)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, rows, 1)
	row := rows[0]
	assert.Equal(t, 1, row.ID)
	assert.Equal(t, 42, row.UserID)
	assert.Equal(t, "gpt-4o", row.ModelName)
	assert.Equal(t, "alice", row.Username)
	assert.Equal(t, 7, row.ChannelID)
	assert.Equal(t, "vip", row.Group)
	assert.Equal(t, "req-2", row.RequestID)
	assert.Equal(t, "up-2", row.UpstreamRequestID)
	assert.Equal(t, 120, row.TotalTokens)
	assert.Equal(t, int64(300), row.ActualCostMicros)
	assert.Equal(t, "USD", row.Currency)

	// Pagination follows the same zero-based offset contract as GetAllLogs.
	rows, total, err = GetCostStatisticsDetails(CostStatisticsDetailFilter{CostStatisticsFilter: CostStatisticsFilter{StartTimestamp: 100, EndTimestamp: 125}}, 1, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	require.Len(t, rows, 1)
	assert.Equal(t, 2, rows[0].ID)
}

func exerciseCostStatisticsDetailsQuery(t *testing.T, db *gorm.DB, databaseType common.DatabaseType) {
	t.Helper()
	previousLogDB := LOG_DB
	previousMainType, previousLogType := common.MainDatabaseType(), common.LogDatabaseType()
	LOG_DB = db
	common.SetDatabaseTypes(databaseType, databaseType)
	initCol()
	t.Cleanup(func() {
		LOG_DB = previousLogDB
		common.SetDatabaseTypes(previousMainType, previousLogType)
		initCol()
	})

	require.NoError(t, db.AutoMigrate(&Log{}))
	requestID := "cost-details-" + strings.ReplaceAll(t.Name(), "/", "-")
	row := &Log{
		CreatedAt:        200,
		UserId:           84,
		Type:             LogTypeConsume,
		BillingEvent:     "request",
		ModelName:        "dialect-model",
		Username:         "dialect-user",
		ChannelId:        17,
		Group:            "dialect-group",
		RequestId:        requestID,
		CostCurrency:     "USD",
		PromptTokens:     4,
		CompletionTokens: 3,
		ActualCostMicros: 5,
		RevenueMicros:    9,
	}
	require.NoError(t, db.Create(row).Error)
	t.Cleanup(func() { db.Where("request_id = ?", requestID).Delete(&Log{}) })

	rows, total, err := GetCostStatisticsDetails(CostStatisticsDetailFilter{
		CostStatisticsFilter: CostStatisticsFilter{StartTimestamp: 199, EndTimestamp: 201, Group: "dialect-group"},
		Currency:             "USD",
		RequestID:            requestID,
	}, 0, 10)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, rows, 1)
	assert.Equal(t, "dialect-group", rows[0].Group)
	assert.Equal(t, 84, rows[0].UserID)
	assert.Equal(t, 7, rows[0].TotalTokens)
}

func TestGetCostStatisticsDetailsMySQL(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("TEST_MYSQL_DSN"))
	if dsn == "" {
		t.Skip("TEST_MYSQL_DSN is not configured")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	exerciseCostStatisticsDetailsQuery(t, db, common.DatabaseTypeMySQL)
}

func TestGetCostStatisticsDetailsPostgreSQL(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("TEST_POSTGRES_DSN"))
	if dsn == "" {
		t.Skip("TEST_POSTGRES_DSN is not configured")
	}
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: dsn, PreferSimpleProtocol: true}), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	exerciseCostStatisticsDetailsQuery(t, db, common.DatabaseTypePostgreSQL)
}
