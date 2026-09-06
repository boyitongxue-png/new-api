package controller

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type costStatisticsAPIResponse struct {
	Success bool                 `json:"success"`
	Data    model.CostStatistics `json:"data"`
}

type costStatisticsLogsAPIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		Page     int                          `json:"page"`
		PageSize int                          `json:"page_size"`
		Total    int                          `json:"total"`
		Items    []model.CostStatisticsDetail `json:"items"`
	} `json:"data"`
}

func setupCostStatisticsControllerTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	initModelListColumnNames(t)

	previousDB := model.DB
	previousLogDB := model.LOG_DB
	previousMainType := common.MainDatabaseType()
	previousLogType := common.LogDatabaseType()
	previousRedis := common.RedisEnabled
	previousSecret := common.SessionSecret
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.UserSession{}, &model.Log{}))

	model.DB = db
	model.LOG_DB = db
	common.SetDatabaseTypes(common.DatabaseTypeSQLite, common.DatabaseTypeSQLite)
	common.RedisEnabled = false
	common.SessionSecret = "cost-statistics-controller-test-secret"
	gin.SetMode(gin.TestMode)
	t.Cleanup(func() {
		model.DB = previousDB
		model.LOG_DB = previousLogDB
		common.SetDatabaseTypes(previousMainType, previousLogType)
		common.RedisEnabled = previousRedis
		common.SessionSecret = previousSecret
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	return db
}

func createCostStatisticsAuthUser(t *testing.T, role int, username string) *service.AuthBundle {
	t.Helper()
	user := &model.User{
		Username:    username,
		Password:    "password-placeholder",
		Role:        role,
		Status:      common.UserStatusEnabled,
		Group:       "default",
		AffCode:     username,
		AuthVersion: 1,
	}
	require.NoError(t, model.DB.Create(user).Error)
	bundle, err := service.CreateLoginSession(user.Id, "password", "127.0.0.1", "cost-statistics-test")
	require.NoError(t, err)
	return bundle
}

func serveCostStatisticsRequest(t *testing.T, path string, accessToken string) *httptest.ResponseRecorder {
	t.Helper()
	router := gin.New()
	router.GET("/api/cost-statistics", middleware.AdminAuth(), GetCostStatistics)
	router.GET("/api/cost-statistics/logs", middleware.AdminAuth(), GetCostStatisticsLogs)
	router.GET("/api/cost-statistics/export", middleware.AdminAuth(), ExportCostStatistics)

	request := httptest.NewRequest(http.MethodGet, path, nil)
	request.Header.Set("Authorization", "Bearer "+accessToken)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func TestCostStatisticsAdminCanAccessAndNonAdminIsRejected(t *testing.T) {
	db := setupCostStatisticsControllerTestDB(t)
	require.NoError(t, db.Create(&model.Log{
		ModelName: "auth-model", Username: "alice", ChannelId: 7, CreatedAt: 100,
		Type: model.LogTypeConsume, BillingEvent: "request", RevenueMicros: 100,
	}).Error)

	admin := createCostStatisticsAuthUser(t, common.RoleAdminUser, "cost-admin")
	adminResponse := serveCostStatisticsRequest(t, "/api/cost-statistics?start_timestamp=1&end_timestamp=200", admin.AccessToken)
	assert.Equal(t, http.StatusOK, adminResponse.Code)
	var adminPayload costStatisticsAPIResponse
	require.NoError(t, common.Unmarshal(adminResponse.Body.Bytes(), &adminPayload))
	assert.True(t, adminPayload.Success)
	require.Len(t, adminPayload.Data.Rows, 1)

	commonUser := createCostStatisticsAuthUser(t, common.RoleCommonUser, "cost-user")
	userResponse := serveCostStatisticsRequest(t, "/api/cost-statistics?start_timestamp=1&end_timestamp=200", commonUser.AccessToken)
	assert.Equal(t, http.StatusForbidden, userResponse.Code)
	assert.Contains(t, userResponse.Body.String(), "AUTH_INSUFFICIENT_PRIVILEGE")
}

func TestCostStatisticsControllerPassesFiltersToModel(t *testing.T) {
	db := setupCostStatisticsControllerTestDB(t)
	logs := []*model.Log{
		{ModelName: "filtered-model", Username: "alice", ChannelId: 7, CreatedAt: 100, Type: model.LogTypeConsume, BillingEvent: "request", RevenueMicros: 1000},
		{ModelName: "filtered-model", Username: "bob", ChannelId: 7, CreatedAt: 100, Type: model.LogTypeConsume, BillingEvent: "request", RevenueMicros: 2000},
		{ModelName: "other-model", Username: "alice", ChannelId: 7, CreatedAt: 100, Type: model.LogTypeConsume, BillingEvent: "request", RevenueMicros: 3000},
		{ModelName: "filtered-model", Username: "alice", ChannelId: 8, CreatedAt: 100, Type: model.LogTypeConsume, BillingEvent: "request", RevenueMicros: 4000},
		{ModelName: "filtered-model", Username: "alice", ChannelId: 7, CreatedAt: 99, Type: model.LogTypeConsume, BillingEvent: "request", RevenueMicros: 5000},
	}
	for _, log := range logs {
		require.NoError(t, db.Create(log).Error)
	}

	stats, err := model.GetCostStatistics(model.CostStatisticsFilter{
		StartTimestamp: 100,
		EndTimestamp:   100,
		ModelName:      "filtered-model",
		Username:       "alice",
		ChannelID:      7,
	})
	require.NoError(t, err)
	require.Len(t, stats.Rows, 1)
	assert.Equal(t, int64(1000), stats.Rows[0].RevenueMicros)

	// The controller uses the same query parser for both endpoints; the API
	// contract is checked through the actual authenticated route below.
	admin := createCostStatisticsAuthUser(t, common.RoleAdminUser, "filter-admin")
	response := serveCostStatisticsRequest(t, "/api/cost-statistics?start_timestamp=100&end_timestamp=100&model_name=filtered-model&username=alice&channel=7", admin.AccessToken)
	assert.Equal(t, http.StatusOK, response.Code)
	var payload costStatisticsAPIResponse
	require.NoError(t, common.Unmarshal(response.Body.Bytes(), &payload))
	require.Len(t, payload.Data.Rows, 1)
	assert.Equal(t, int64(1000), payload.Data.Rows[0].RevenueMicros)
}

func TestExportCostStatisticsEscapesFormulaLikeCells(t *testing.T) {
	db := setupCostStatisticsControllerTestDB(t)
	for index, modelName := range []string{"=SUM(1,1)", "+cmd", "-cmd", "@cmd"} {
		require.NoError(t, db.Create(&model.Log{
			ModelName: modelName, ChannelId: index + 1, CreatedAt: 100,
			Type: model.LogTypeConsume, BillingEvent: "request", CostCurrency: "USD",
			RevenueMicros: 100,
		}).Error)
	}

	admin := createCostStatisticsAuthUser(t, common.RoleAdminUser, "export-admin")
	response := serveCostStatisticsRequest(t, "/api/cost-statistics/export?start_timestamp=100&end_timestamp=100", admin.AccessToken)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "text/csv; charset=utf-8", response.Header().Get("Content-Type"))
	assert.Contains(t, response.Header().Get("Content-Disposition"), "cost-statistics.csv")

	records, err := csv.NewReader(strings.NewReader(response.Body.String())).ReadAll()
	require.NoError(t, err)
	require.Len(t, records, 5)
	seen := make(map[string]bool, 4)
	for _, record := range records[1:] {
		require.GreaterOrEqual(t, len(record), 1)
		seen[record[0]] = true
		assert.True(t, strings.HasPrefix(record[0], "'"), "formula-like model %q must be prefixed before CSV export", record[0])
	}
	for _, modelName := range []string{"=SUM(1,1)", "+cmd", "-cmd", "@cmd"} {
		assert.True(t, seen["'"+modelName])
	}
}

func TestCostStatisticsLogsAPIIsAdminProtectedFilteredAndRedacted(t *testing.T) {
	db := setupCostStatisticsControllerTestDB(t)
	require.NoError(t, db.Create(&model.Log{
		CreatedAt:         100,
		UserId:            42,
		Type:              model.LogTypeConsume,
		BillingEvent:      "request",
		ModelName:         "gpt-4o",
		Username:          "alice",
		ChannelId:         7,
		Group:             "vip",
		RequestId:         "req-safe",
		UpstreamRequestId: "up-safe",
		PromptTokens:      10,
		CompletionTokens:  5,
		ActualCostMicros:  12,
		RevenueMicros:     30,
		CostCurrency:      "USD",
		UsageAvailable:    true,
		Content:           "sensitive prompt",
		Other:             `{"private":"sensitive metadata"}`,
		Ip:                "192.0.2.20",
		TokenId:           123,
		TokenName:         "private-token",
	}).Error)
	require.NoError(t, db.Create(&model.Log{
		CreatedAt: 101, Type: model.LogTypeConsume, BillingEvent: "request",
		ModelName: "gpt-4o", Username: "bob", ChannelId: 8, Group: "default",
		RequestId: "req-other", CostCurrency: "EUR",
	}).Error)

	admin := createCostStatisticsAuthUser(t, common.RoleAdminUser, "details-admin")
	response := serveCostStatisticsRequest(t, "/api/cost-statistics/logs?p=1&page_size=10&model_name=gpt-4o&username=alice&channel=7&group=vip&currency=USD&request_id=req-safe", admin.AccessToken)
	require.Equal(t, http.StatusOK, response.Code)
	var payload costStatisticsLogsAPIResponse
	require.NoError(t, common.Unmarshal(response.Body.Bytes(), &payload))
	require.True(t, payload.Success, payload.Message)
	require.Equal(t, 1, payload.Data.Page)
	require.Equal(t, 10, payload.Data.PageSize)
	require.Equal(t, 1, payload.Data.Total)
	require.Len(t, payload.Data.Items, 1)
	assert.Equal(t, "req-safe", payload.Data.Items[0].RequestID)
	assert.Equal(t, 42, payload.Data.Items[0].UserID)
	assert.Equal(t, 15, payload.Data.Items[0].TotalTokens)
	assert.Equal(t, int64(12), payload.Data.Items[0].ActualCostMicros)

	body := strings.ToLower(response.Body.String())
	for _, forbidden := range []string{"sensitive prompt", "sensitive metadata", "192.0.2.20", "private-token", "token_id", "token_name", `"content"`, `"other"`, `"ip"`} {
		assert.NotContains(t, body, strings.ToLower(forbidden))
	}

	commonUser := createCostStatisticsAuthUser(t, common.RoleCommonUser, "details-user")
	forbidden := serveCostStatisticsRequest(t, "/api/cost-statistics/logs?request_id=req-safe", commonUser.AccessToken)
	assert.Equal(t, http.StatusForbidden, forbidden.Code)
	assert.Contains(t, forbidden.Body.String(), "AUTH_INSUFFICIENT_PRIVILEGE")
}
