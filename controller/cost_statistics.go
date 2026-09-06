package controller

import (
	"encoding/csv"
	"net/http"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

func costStatisticsFilter(c *gin.Context) model.CostStatisticsFilter {
	start, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	end, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	channel, _ := strconv.Atoi(c.Query("channel"))
	return model.CostStatisticsFilter{
		StartTimestamp: start,
		EndTimestamp:   end,
		ModelName:      c.Query("model_name"),
		Username:       c.Query("username"),
		ChannelID:      channel,
		Group:          c.Query("group"),
	}
}

func costStatisticsDetailFilter(c *gin.Context) model.CostStatisticsDetailFilter {
	return model.CostStatisticsDetailFilter{
		CostStatisticsFilter: costStatisticsFilter(c),
		Currency:             c.Query("currency"),
		RequestID:            c.Query("request_id"),
	}
}

func GetCostStatistics(c *gin.Context) {
	stats, err := model.GetCostStatistics(costStatisticsFilter(c))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": stats})
}

func GetCostStatisticsLogs(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	if pageInfo.GetPageSize() <= 0 {
		pageInfo.PageSize = common.ItemsPerPage
	}
	details, total, err := model.GetCostStatisticsDetails(costStatisticsDetailFilter(c), pageInfo.GetStartIdx(), pageInfo.GetPageSize())
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(details)
	common.ApiSuccess(c, pageInfo)
}

func csvSafe(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return value
	}
	if strings.ContainsRune("=+-@", rune(value[0])) {
		return "'" + value
	}
	return value
}

func ExportCostStatistics(c *gin.Context) {
	stats, err := model.GetCostStatistics(costStatisticsFilter(c))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="cost-statistics.csv"`)
	writer := csv.NewWriter(c.Writer)
	_ = writer.Write([]string{
		"model", "channel_id", "currency", "request_count", "success_count", "failure_count",
		"prompt_tokens", "completion_tokens", "cached_tokens", "cache_creation_tokens",
		"image_tokens", "audio_tokens", "audio_input_tokens", "audio_output_tokens", "image_count",
		"audio_seconds", "audio_output_seconds", "video_seconds", "total_tokens",
		"actual_cost_micros", "revenue_micros", "gross_profit_micros", "profit_rate", "usage_missing_count",
	})
	for _, row := range stats.Rows {
		_ = writer.Write([]string{
			csvSafe(row.ModelName), strconv.Itoa(row.ChannelID), csvSafe(row.Currency),
			strconv.FormatInt(row.RequestCount, 10), strconv.FormatInt(row.SuccessCount, 10), strconv.FormatInt(row.FailureCount, 10),
			strconv.FormatInt(row.PromptTokens, 10), strconv.FormatInt(row.CompletionTokens, 10), strconv.FormatInt(row.CachedTokens, 10), strconv.FormatInt(row.CacheCreationTokens, 10),
			strconv.FormatInt(row.ImageTokens, 10), strconv.FormatInt(row.AudioTokens, 10), strconv.FormatInt(row.AudioInputTokens, 10), strconv.FormatInt(row.AudioOutputTokens, 10), strconv.FormatInt(row.ImageCount, 10),
			strconv.FormatInt(row.AudioSeconds, 10), strconv.FormatInt(row.AudioOutputSeconds, 10), strconv.FormatInt(row.VideoSeconds, 10), strconv.FormatInt(row.TotalTokens, 10),
			strconv.FormatInt(row.ActualCostMicros, 10), strconv.FormatInt(row.RevenueMicros, 10), strconv.FormatInt(row.GrossProfitMicros, 10), strconv.FormatFloat(row.ProfitRate, 'f', 8, 64), strconv.FormatInt(row.UsageMissingCount, 10),
		})
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return
	}
}
