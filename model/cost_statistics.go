package model

import (
	"errors"
	"time"

	"github.com/QuantumNous/new-api/common"
)

type CostStatisticsFilter struct {
	StartTimestamp int64
	EndTimestamp   int64
	ModelName      string
	Username       string
	ChannelID      int
	Group          string
}

type CostStatisticsDetailFilter struct {
	CostStatisticsFilter
	Currency  string
	RequestID string
}

// CostStatisticsDetail is the safe, accounting-focused projection used by the
// cost report drill-down. It intentionally does not expose Log.Content,
// Log.Other, IP addresses, token secrets, or token identifiers.
type CostStatisticsDetail struct {
	ID                  int     `json:"id"`
	UserID              int     `json:"user_id"`
	CreatedAt           int64   `json:"created_at"`
	Type                int     `json:"type"`
	BillingEvent        string  `json:"billing_event"`
	ModelName           string  `json:"model_name"`
	Username            string  `json:"username"`
	ChannelID           int     `json:"channel_id"`
	Group               string  `json:"group"`
	RequestID           string  `json:"request_id,omitempty"`
	UpstreamRequestID   string  `json:"upstream_request_id,omitempty"`
	Quota               int     `json:"quota"`
	PromptTokens        int     `json:"prompt_tokens"`
	CompletionTokens    int     `json:"completion_tokens"`
	CachedTokens        int     `json:"cached_tokens"`
	CacheCreationTokens int     `json:"cache_creation_tokens"`
	ImageTokens         int     `json:"image_tokens"`
	AudioTokens         int     `json:"audio_tokens"`
	AudioInputTokens    int     `json:"audio_input_tokens"`
	AudioOutputTokens   int     `json:"audio_output_tokens"`
	ImageCount          int     `json:"image_count"`
	AudioSeconds        int     `json:"audio_seconds"`
	AudioOutputSeconds  int     `json:"audio_output_seconds"`
	VideoSeconds        int     `json:"video_seconds"`
	TotalTokens         int     `json:"total_tokens"`
	ActualCostMicros    int64   `json:"actual_cost_micros"`
	RevenueMicros       int64   `json:"revenue_micros"`
	GrossProfitMicros   int64   `json:"gross_profit_micros"`
	ProfitRate          float64 `json:"profit_rate"`
	Currency            string  `json:"currency"`
	UsageAvailable      bool    `json:"usage_available"`
}

type ModelCostStatistic struct {
	ModelName           string  `json:"model_name"`
	ChannelID           int     `json:"channel_id"`
	Currency            string  `json:"currency"`
	RequestCount        int64   `json:"request_count"`
	SuccessCount        int64   `json:"success_count"`
	FailureCount        int64   `json:"failure_count"`
	PromptTokens        int64   `json:"prompt_tokens"`
	CompletionTokens    int64   `json:"completion_tokens"`
	CachedTokens        int64   `json:"cached_tokens"`
	CacheCreationTokens int64   `json:"cache_creation_tokens"`
	ImageTokens         int64   `json:"image_tokens"`
	AudioTokens         int64   `json:"audio_tokens"`
	AudioInputTokens    int64   `json:"audio_input_tokens"`
	AudioOutputTokens   int64   `json:"audio_output_tokens"`
	ImageCount          int64   `json:"image_count"`
	AudioSeconds        int64   `json:"audio_seconds"`
	AudioOutputSeconds  int64   `json:"audio_output_seconds"`
	VideoSeconds        int64   `json:"video_seconds"`
	TotalTokens         int64   `json:"total_tokens"`
	ActualCostMicros    int64   `json:"actual_cost_micros"`
	RevenueMicros       int64   `json:"revenue_micros"`
	GrossProfitMicros   int64   `json:"gross_profit_micros"`
	ProfitRate          float64 `json:"profit_rate"`
	UsageMissingCount   int64   `json:"usage_missing_count"`
}

type CostStatistics struct {
	Rows              []*ModelCostStatistic `json:"rows"`
	Currency          string                `json:"currency"`
	MixedCurrency     bool                  `json:"mixed_currency"`
	RequestCount      int64                 `json:"request_count"`
	SuccessCount      int64                 `json:"success_count"`
	FailureCount      int64                 `json:"failure_count"`
	PromptTokens      int64                 `json:"prompt_tokens"`
	CompletionTokens  int64                 `json:"completion_tokens"`
	CachedTokens      int64                 `json:"cached_tokens"`
	AudioTokens       int64                 `json:"audio_tokens"`
	AudioInputTokens  int64                 `json:"audio_input_tokens"`
	AudioOutputTokens int64                 `json:"audio_output_tokens"`
	ImageCount        int64                 `json:"image_count"`
	AudioSeconds      int64                 `json:"audio_seconds"`
	VideoSeconds      int64                 `json:"video_seconds"`
	TotalTokens       int64                 `json:"total_tokens"`
	ActualCostMicros  int64                 `json:"actual_cost_micros"`
	RevenueMicros     int64                 `json:"revenue_micros"`
	GrossProfitMicros int64                 `json:"gross_profit_micros"`
	ProfitRate        float64               `json:"profit_rate"`
}

func normalizeCostStatisticsFilter(filter CostStatisticsFilter) CostStatisticsFilter {
	if filter.StartTimestamp == 0 {
		filter.StartTimestamp = time.Now().Add(-30 * 24 * time.Hour).Unix()
	}
	if filter.EndTimestamp == 0 {
		filter.EndTimestamp = time.Now().Unix()
	}
	return filter
}

func GetCostStatistics(filter CostStatisticsFilter) (*CostStatistics, error) {
	if LOG_DB == nil {
		return nil, errors.New("log database is not initialized")
	}
	filter = normalizeCostStatisticsFilter(filter)
	query := LOG_DB.Table("logs").Where("type IN (?, ?, ?)", LogTypeConsume, LogTypeError, LogTypeRefund)
	query = query.Where("created_at >= ? AND created_at <= ?", filter.StartTimestamp, filter.EndTimestamp)
	if filter.ModelName != "" {
		query = query.Where("model_name = ?", filter.ModelName)
	}
	if filter.Username != "" {
		query = query.Where("username = ?", filter.Username)
	}
	if filter.ChannelID != 0 {
		query = query.Where("channel_id = ?", filter.ChannelID)
	}
	if filter.Group != "" {
		query = query.Where(logGroupCol+" = ?", filter.Group)
	}

	var rows []*ModelCostStatistic
	selectSQL := `model_name, channel_id,
COALESCE(NULLIF(cost_currency, ''), 'USD') AS currency,
SUM(CASE WHEN type = ? AND (billing_event = 'request' OR billing_event = '') THEN 1 ELSE 0 END) AS request_count,
SUM(CASE WHEN type = ? AND (billing_event = 'request' OR billing_event = '') THEN 1 ELSE 0 END) AS success_count,
SUM(CASE WHEN type = ? THEN 1 ELSE 0 END) AS failure_count,
COALESCE(SUM(prompt_tokens), 0) AS prompt_tokens,
COALESCE(SUM(completion_tokens), 0) AS completion_tokens,
COALESCE(SUM(cached_tokens), 0) AS cached_tokens,
COALESCE(SUM(cache_creation_tokens), 0) AS cache_creation_tokens,
COALESCE(SUM(image_tokens), 0) AS image_tokens,
COALESCE(SUM(audio_tokens), 0) AS audio_tokens,
COALESCE(SUM(audio_input_tokens), 0) AS audio_input_tokens,
COALESCE(SUM(audio_output_tokens), 0) AS audio_output_tokens,
COALESCE(SUM(image_count), 0) AS image_count,
COALESCE(SUM(audio_seconds), 0) AS audio_seconds,
COALESCE(SUM(audio_output_seconds), 0) AS audio_output_seconds,
COALESCE(SUM(video_seconds), 0) AS video_seconds,
COALESCE(SUM(prompt_tokens + completion_tokens), 0) AS total_tokens,
COALESCE(SUM(CASE WHEN type = ? THEN actual_cost_micros ELSE 0 END), 0) AS actual_cost_micros,
COALESCE(SUM(CASE WHEN type IN (?, ?) THEN revenue_micros ELSE 0 END), 0) AS revenue_micros,
COALESCE(SUM(CASE WHEN type = ? AND usage_available = ? THEN 1 ELSE 0 END), 0) AS usage_missing_count`
	err := query.Select(selectSQL, LogTypeConsume, LogTypeConsume, LogTypeError, LogTypeConsume, LogTypeConsume, LogTypeRefund, LogTypeConsume, false).Group("model_name, channel_id, COALESCE(NULLIF(cost_currency, ''), 'USD')").Order("actual_cost_micros DESC").Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := &CostStatistics{Rows: rows}
	for _, row := range rows {
		if result.Currency == "" {
			result.Currency = row.Currency
		} else if result.Currency != row.Currency {
			result.MixedCurrency = true
		}
		row.GrossProfitMicros = row.RevenueMicros - row.ActualCostMicros
		if row.RevenueMicros > 0 {
			row.ProfitRate = float64(row.GrossProfitMicros) / float64(row.RevenueMicros)
		}
		result.RequestCount += row.RequestCount
		result.SuccessCount += row.SuccessCount
		result.FailureCount += row.FailureCount
		result.PromptTokens += row.PromptTokens
		result.CompletionTokens += row.CompletionTokens
		result.CachedTokens += row.CachedTokens
		result.AudioTokens += row.AudioTokens
		result.AudioInputTokens += row.AudioInputTokens
		result.AudioOutputTokens += row.AudioOutputTokens
		result.ImageCount += row.ImageCount
		result.AudioSeconds += row.AudioSeconds
		result.VideoSeconds += row.VideoSeconds
		result.TotalTokens += row.TotalTokens
		if !result.MixedCurrency {
			result.ActualCostMicros += row.ActualCostMicros
			result.RevenueMicros += row.RevenueMicros
			result.GrossProfitMicros += row.GrossProfitMicros
		}
	}
	if result.MixedCurrency {
		result.Currency = ""
		result.ActualCostMicros = 0
		result.RevenueMicros = 0
		result.GrossProfitMicros = 0
		result.ProfitRate = 0
	} else if result.RevenueMicros > 0 {
		result.ProfitRate = float64(result.GrossProfitMicros) / float64(result.RevenueMicros)
	}
	return result, nil
}

// GetCostStatisticsDetails returns the accounting-related log rows behind a
// cost-statistics summary row. The query uses an explicit projection instead
// of loading Log so sensitive request content and operational metadata cannot
// accidentally become part of this endpoint's response.
func GetCostStatisticsDetails(filter CostStatisticsDetailFilter, startIdx int, num int) (details []*CostStatisticsDetail, total int64, err error) {
	if LOG_DB == nil {
		return nil, 0, errors.New("log database is not initialized")
	}
	base := filter.CostStatisticsFilter
	query := LOG_DB.Table("logs").Where("logs.type IN (?, ?, ?)", LogTypeConsume, LogTypeError, LogTypeRefund)
	if base.StartTimestamp != 0 {
		query = query.Where("logs.created_at >= ?", base.StartTimestamp)
	}
	if base.EndTimestamp != 0 {
		query = query.Where("logs.created_at <= ?", base.EndTimestamp)
	}
	if base.ModelName != "" {
		if query, err = applyExplicitLogTextFilter(query, "logs.model_name", base.ModelName); err != nil {
			return nil, 0, err
		}
	}
	if base.Username != "" {
		if query, err = applyExplicitLogTextFilter(query, "logs.username", base.Username); err != nil {
			return nil, 0, err
		}
	}
	if base.ChannelID != 0 {
		query = query.Where("logs.channel_id = ?", base.ChannelID)
	}
	if base.Group != "" {
		query = query.Where("logs."+logGroupCol+" = ?", base.Group)
	}
	if filter.Currency != "" {
		query = query.Where("COALESCE(NULLIF(logs.cost_currency, ''), 'USD') = ?", filter.Currency)
	}
	if filter.RequestID != "" {
		query = query.Where("logs.request_id = ?", filter.RequestID)
	}

	if err = query.Model(&Log{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	selectSQL := `logs.id, logs.user_id, logs.created_at, logs.type, logs.billing_event,
logs.model_name, logs.username, logs.channel_id, logs.` + logGroupCol + ` AS ` + logGroupCol + `,
logs.request_id, logs.upstream_request_id, logs.quota,
logs.prompt_tokens, logs.completion_tokens, logs.cached_tokens, logs.cache_creation_tokens,
logs.image_tokens, logs.audio_tokens, logs.audio_input_tokens, logs.audio_output_tokens,
logs.image_count, logs.audio_seconds, logs.audio_output_seconds, logs.video_seconds,
COALESCE(logs.prompt_tokens, 0) + COALESCE(logs.completion_tokens, 0) AS total_tokens,
logs.actual_cost_micros, logs.revenue_micros,
logs.revenue_micros - logs.actual_cost_micros AS gross_profit_micros,
CASE WHEN logs.revenue_micros > 0 THEN (logs.revenue_micros - logs.actual_cost_micros) * 1.0 / logs.revenue_micros ELSE 0 END AS profit_rate,
COALESCE(NULLIF(logs.cost_currency, ''), 'USD') AS currency, logs.usage_available`
	order := "logs.created_at desc, logs.id desc"
	if common.UsingLogDatabase(common.DatabaseTypeClickHouse) {
		order = clickHouseLogOrder("logs.")
	}
	err = query.Select(selectSQL).Order(order).Limit(num).Offset(startIdx).Find(&details).Error
	return details, total, err
}
