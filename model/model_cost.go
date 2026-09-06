package model

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/shopspring/decimal"
)

// ModelCost describes upstream cost in USD. Token prices use USD per 1M
// tokens; unit prices use USD per image/second/request as named by the field.
type ModelCost struct {
	Currency              string  `json:"currency"`
	InputPerMillion       float64 `json:"input_per_1m"`
	OutputPerMillion      float64 `json:"output_per_1m"`
	CacheReadPerMillion   float64 `json:"cache_read_per_1m"`
	CacheWritePerMillion  float64 `json:"cache_write_per_1m"`
	ImagePerUnit          float64 `json:"image_per_unit"`
	ImageTokenPerMillion  float64 `json:"image_token_per_1m"`
	AudioInputPerMillion  float64 `json:"audio_input_per_1m"`
	AudioOutputPerMillion float64 `json:"audio_output_per_1m"`
	AudioInputPerSecond   float64 `json:"audio_input_per_second"`
	AudioOutputPerSecond  float64 `json:"audio_output_per_second"`
	VideoPerSecond        float64 `json:"video_per_second"`
	RequestFee            float64 `json:"request_fee"`
	Enabled               bool    `json:"enabled"`
}

// ModelCostConfig is keyed by model name. The optional channel:<id>:<model>
// key has higher priority than the model key, followed by default.
type ModelCostConfig map[string]ModelCost

const modelCostOptionKey = "ModelCost"

const maxModelCostCurrencyLength = 8

func GetModelCostConfig() ModelCostConfig {
	common.OptionMapRWMutex.RLock()
	raw := common.OptionMap[modelCostOptionKey]
	common.OptionMapRWMutex.RUnlock()
	if strings.TrimSpace(raw) == "" {
		return ModelCostConfig{}
	}
	var config ModelCostConfig
	if err := common.UnmarshalJsonStr(raw, &config); err != nil {
		common.SysError("invalid ModelCost option: " + err.Error())
		return ModelCostConfig{}
	}
	return config
}

func ValidateModelCostJSON(raw string) error {
	var config ModelCostConfig
	if err := common.UnmarshalJsonStr(raw, &config); err != nil {
		return fmt.Errorf("ModelCost must be a JSON object: %w", err)
	}
	for key, cost := range config {
		if strings.TrimSpace(key) == "" {
			return fmt.Errorf("ModelCost contains an empty model key")
		}
		currency := strings.TrimSpace(cost.Currency)
		if currency == "" {
			currency = "USD"
		}
		if len(currency) > maxModelCostCurrencyLength {
			return fmt.Errorf("ModelCost.%s.currency must be at most %d characters", key, maxModelCostCurrencyLength)
		}
		if err := validateModelCostValues(key, cost); err != nil {
			return err
		}
	}
	return nil
}

func validateModelCostValues(key string, cost ModelCost) error {
	for name, value := range map[string]float64{
		"input_per_1m":            cost.InputPerMillion,
		"output_per_1m":           cost.OutputPerMillion,
		"cache_read_per_1m":       cost.CacheReadPerMillion,
		"cache_write_per_1m":      cost.CacheWritePerMillion,
		"image_per_unit":          cost.ImagePerUnit,
		"image_token_per_1m":      cost.ImageTokenPerMillion,
		"audio_input_per_1m":      cost.AudioInputPerMillion,
		"audio_output_per_1m":     cost.AudioOutputPerMillion,
		"audio_input_per_second":  cost.AudioInputPerSecond,
		"audio_output_per_second": cost.AudioOutputPerSecond,
		"video_per_second":        cost.VideoPerSecond,
		"request_fee":             cost.RequestFee,
	} {
		if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || value > 1_000_000_000 {
			return fmt.Errorf("ModelCost.%s.%s must be finite and between 0 and 1000000000", key, name)
		}
	}
	return nil
}

func normalizeModelCost(key string, cost ModelCost) (ModelCost, bool) {
	cost.Currency = strings.TrimSpace(cost.Currency)
	if cost.Currency == "" {
		cost.Currency = "USD"
	}
	if len(cost.Currency) > maxModelCostCurrencyLength {
		common.SysError(fmt.Sprintf("invalid ModelCost.%s currency: exceeds %d characters", key, maxModelCostCurrencyLength))
		return ModelCost{}, false
	}
	if err := validateModelCostValues(key, cost); err != nil {
		common.SysError(err.Error())
		return ModelCost{}, false
	}
	return cost, true
}

func resolveModelCost(modelName string, channelID int) (ModelCost, string) {
	config := GetModelCostConfig()
	keys := []string{}
	if channelID > 0 {
		keys = append(keys, "channel:"+strconv.Itoa(channelID)+":"+modelName)
	}
	keys = append(keys, modelName, "default")
	for _, key := range keys {
		if cost, ok := config[key]; ok && cost.Enabled {
			if normalized, valid := normalizeModelCost(key, cost); valid {
				return normalized, key
			}
		}
	}
	return ModelCost{}, ""
}

func costMicros(price float64, units int64, perMillion bool) int64 {
	if math.IsNaN(price) || math.IsInf(price, 0) || price <= 0 || units <= 0 {
		return 0
	}
	value := decimal.NewFromFloat(price).Mul(decimal.NewFromInt(units))
	if perMillion {
		value = value.Div(decimal.NewFromInt(1_000_000))
	}
	value = value.Mul(decimal.NewFromInt(1_000_000))
	max := decimal.NewFromInt(math.MaxInt64)
	if value.GreaterThan(max) {
		common.SysError("model cost saturated at int64 maximum")
		return math.MaxInt64
	}
	return value.Round(0).IntPart()
}

func addCostMicros(total int64, value int64) int64 {
	if value <= 0 {
		return total
	}
	if total >= math.MaxInt64-value {
		if value > 0 && total < math.MaxInt64 {
			common.SysError("model cost sum saturated at int64 maximum")
		}
		return math.MaxInt64
	}
	return total + value
}

func revenueMicros(quota int) int64 {
	if quota == 0 || common.QuotaPerUnit <= 0 {
		return 0
	}
	value := decimal.NewFromInt(int64(quota)).Mul(decimal.NewFromInt(1_000_000)).Div(decimal.NewFromFloat(common.QuotaPerUnit)).Round(0)
	if value.GreaterThan(decimal.NewFromInt(math.MaxInt64)) {
		return math.MaxInt64
	}
	if value.LessThan(decimal.NewFromInt(math.MinInt64)) {
		return math.MinInt64
	}
	return value.IntPart()
}

// CostAccounting is the immutable per-request accounting snapshot stored in a
// consume/refund log. Values are integer micro-USD to keep reports stable when
// current pricing or QuotaPerUnit changes.
type CostAccounting struct {
	ActualCostMicros int64  `json:"actual_cost_micros"`
	RevenueMicros    int64  `json:"revenue_micros"`
	Currency         string `json:"currency"`
	Source           string `json:"source"`
	UsageAvailable   bool   `json:"usage_available"`
}

// CostUsage contains the billable dimensions captured from one upstream response.
// Audio input/output tokens are separate so an output price can never be silently
// applied to the wrong unit. Audio/video seconds are integer seconds in the log.
type CostUsage struct {
	UsageAvailable      bool
	PromptTokens        int64
	CompletionTokens    int64
	CachedTokens        int64
	CacheCreationTokens int64
	ImageTokens         int64
	AudioInputTokens    int64
	AudioOutputTokens   int64
	ImageCount          int64
	AudioSeconds        int64
	AudioOutputSeconds  int64
	VideoSeconds        int64
	IncludeRequestFee   bool
}

// CalculateCostAccounting preserves the original call shape for non-relay
// callers and treats the legacy audioTokens value as input audio tokens.
func CalculateCostAccounting(modelName string, channelID int, quota int, usageAvailable bool, promptTokens, completionTokens, cachedTokens, cacheCreationTokens, imageTokens, audioTokens, imageCount, audioSeconds, videoSeconds int64) CostAccounting {
	return CalculateCostAccountingWithUsage(modelName, channelID, quota, CostUsage{
		UsageAvailable:      usageAvailable,
		PromptTokens:        promptTokens,
		CompletionTokens:    completionTokens,
		CachedTokens:        cachedTokens,
		CacheCreationTokens: cacheCreationTokens,
		ImageTokens:         imageTokens,
		AudioInputTokens:    audioTokens,
		ImageCount:          imageCount,
		AudioSeconds:        audioSeconds,
		VideoSeconds:        videoSeconds,
		IncludeRequestFee:   true,
	})
}

func CalculateCostAccountingWithUsage(modelName string, channelID int, quota int, usage CostUsage) CostAccounting {
	cost, source := resolveModelCost(modelName, channelID)
	accounting := CostAccounting{
		RevenueMicros:  revenueMicros(quota),
		Currency:       cost.Currency,
		Source:         source,
		UsageAvailable: usage.UsageAvailable,
	}
	if source == "" {
		return accounting
	}
	var total int64
	if usage.UsageAvailable {
		// Prompt/completion totals may include modality-specific tokens. Charge
		// those tokens through their dedicated price first, then charge only the
		// remaining text tokens at the generic input/output rates.
		inputTextTokens := usage.PromptTokens - usage.AudioInputTokens - usage.ImageTokens
		if inputTextTokens < 0 {
			inputTextTokens = 0
		}
		outputTextTokens := usage.CompletionTokens - usage.AudioOutputTokens
		if outputTextTokens < 0 {
			outputTextTokens = 0
		}
		for _, item := range []struct {
			price      float64
			units      int64
			perMillion bool
		}{
			{cost.InputPerMillion, inputTextTokens, true},
			{cost.OutputPerMillion, outputTextTokens, true},
			{cost.CacheReadPerMillion, usage.CachedTokens, true},
			{cost.CacheWritePerMillion, usage.CacheCreationTokens, true},
			{cost.ImageTokenPerMillion, usage.ImageTokens, true},
			{cost.ImagePerUnit, usage.ImageCount, false},
			{cost.AudioInputPerMillion, usage.AudioInputTokens, true},
			{cost.AudioOutputPerMillion, usage.AudioOutputTokens, true},
			{cost.AudioInputPerSecond, usage.AudioSeconds, false},
			{cost.AudioOutputPerSecond, usage.AudioOutputSeconds, false},
			{cost.VideoPerSecond, usage.VideoSeconds, false},
		} {
			total = addCostMicros(total, costMicros(item.price, item.units, item.perMillion))
		}
	}
	if usage.IncludeRequestFee && total < math.MaxInt64 {
		total = addCostMicros(total, costMicros(cost.RequestFee, 1, false))
	}
	accounting.ActualCostMicros = total
	return accounting
}
