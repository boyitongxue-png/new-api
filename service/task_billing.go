package service

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/setting/billing_setting"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
	"github.com/QuantumNous/new-api/types"
	"github.com/gin-gonic/gin"
)

// LogTaskConsumption 记录任务消费日志和统计信息（仅记录，不涉及实际扣费）。
// 实际扣费已由 BillingSession（PreConsumeBilling + SettleBilling）完成。
func LogTaskConsumption(c *gin.Context, info *relaycommon.RelayInfo) {
	tokenName := c.GetString("token_name")
	logContent := fmt.Sprintf("操作 %s", info.Action)
	taskBillingMode := billing_setting.GetTaskBillingMode(
		info.OriginModelName,
		common.StringsContains(constant.TaskPricePatches, info.OriginModelName),
	)
	if info.TaskBillingPrice != nil {
		taskBillingMode = info.TaskBillingPrice.Mode
	}
	// 任务按次计费时不展示时长、分辨率等倍率，避免日志与实际计费单位混淆。
	if taskBillingMode == billing_setting.BillingModePerRequest {
		logContent = fmt.Sprintf("%s，按次计费", logContent)
	} else {
		if otherRatios := info.PriceData.OtherRatios(); len(otherRatios) > 0 {
			var contents []string
			for key, ra := range otherRatios {
				if 1.0 != ra {
					contents = append(contents, fmt.Sprintf("%s: %.2f", key, ra))
				}
			}
			if len(contents) > 0 {
				logContent = fmt.Sprintf("%s, 计算参数：%s", logContent, strings.Join(contents, ", "))
			}
		}
	}
	other := make(map[string]interface{})
	other["is_task"] = true
	other["request_path"] = c.Request.URL.Path
	other["model_price"] = info.PriceData.ModelPrice
	if info.PriceData.ModelRatio > 0 {
		other["model_ratio"] = info.PriceData.ModelRatio
	}
	other["group_ratio"] = info.PriceData.GroupRatioInfo.GroupRatio
	if info.PriceData.GroupRatioInfo.HasSpecialRatio {
		other["user_group_ratio"] = info.PriceData.GroupRatioInfo.GroupSpecialRatio
	}
	attachQuotaSaturation(c, info, other)
	model.RecordConsumeLog(c, info.UserId, model.RecordConsumeLogParams{
		ChannelId: info.ChannelId,
		ModelName: info.OriginModelName,
		TokenName: tokenName,
		Quota:     info.PriceData.Quota,
		Content:   logContent,
		TokenId:   info.TokenId,
		Group:     info.UsingGroup,
		Other:     other,
	})
	model.UpdateUserUsedQuotaAndRequestCount(info.UserId, info.PriceData.Quota)
	model.UpdateChannelUsedQuota(info.ChannelId, info.PriceData.Quota)
}

// ---------------------------------------------------------------------------
// 异步任务计费辅助函数
// ---------------------------------------------------------------------------

// resolveTokenKey 通过 TokenId 运行时获取令牌 Key（用于 Redis 缓存操作）。
// 如果令牌已被删除或查询失败，返回空字符串。
func resolveTokenKey(ctx context.Context, tokenId int, taskID string) string {
	token, err := model.GetTokenById(tokenId)
	if err != nil {
		logger.LogWarn(ctx, fmt.Sprintf("获取令牌 key 失败 (tokenId=%d, task=%s): %s", tokenId, taskID, err.Error()))
		return ""
	}
	return token.Key
}

// taskIsSubscription 判断任务是否通过订阅计费。
func taskIsSubscription(task *model.Task) bool {
	return task.PrivateData.BillingSource == BillingSourceSubscription && task.PrivateData.SubscriptionId > 0
}

// taskAdjustFunding 调整任务的资金来源（钱包或订阅），delta > 0 表示扣费，delta < 0 表示退还。
func taskAdjustFunding(task *model.Task, delta int) error {
	if taskIsSubscription(task) {
		return model.PostConsumeUserSubscriptionDelta(task.PrivateData.SubscriptionId, int64(delta))
	}
	if delta > 0 {
		return model.DecreaseUserQuota(task.UserId, delta, false)
	}
	return model.IncreaseUserQuota(task.UserId, -delta, false)
}

// taskAdjustTokenQuota 调整任务的令牌额度，delta > 0 表示扣费，delta < 0 表示退还。
// 需要通过 resolveTokenKey 运行时获取 key（不从 PrivateData 中读取）。
func taskAdjustTokenQuota(ctx context.Context, task *model.Task, delta int) {
	if task.PrivateData.TokenId <= 0 || delta == 0 {
		return
	}
	tokenKey := resolveTokenKey(ctx, task.PrivateData.TokenId, task.TaskID)
	if tokenKey == "" {
		return
	}
	var err error
	if delta > 0 {
		err = model.DecreaseTokenQuota(task.PrivateData.TokenId, tokenKey, delta)
	} else {
		err = model.IncreaseTokenQuota(task.PrivateData.TokenId, tokenKey, -delta)
	}
	if err != nil {
		logger.LogWarn(ctx, fmt.Sprintf("调整令牌额度失败 (delta=%d, task=%s): %s", delta, task.TaskID, err.Error()))
	}
}

func taskUsageNumber(value interface{}) (float64, bool) {
	switch number := value.(type) {
	case float64:
		return number, true
	case float32:
		return float64(number), true
	case int:
		return float64(number), true
	case int64:
		return float64(number), true
	case uint:
		return float64(number), true
	case uint64:
		return float64(number), true
	default:
		return 0, false
	}
}

func normalizeTaskUsageKey(key string) string {
	return strings.NewReplacer("_", "", "-", "").Replace(strings.ToLower(strings.TrimSpace(key)))
}

func readTaskUsageFact(facts map[string]interface{}, aliases []string, limit int) (int64, bool, bool) {
	wanted := make(map[string]struct{}, len(aliases))
	for _, alias := range aliases {
		wanted[normalizeTaskUsageKey(alias)] = struct{}{}
	}
	for key, raw := range facts {
		if _, ok := wanted[normalizeTaskUsageKey(key)]; !ok {
			continue
		}
		number, ok := taskUsageNumber(raw)
		if !ok || math.IsNaN(number) || math.IsInf(number, 0) || number < 0 {
			return 0, true, true
		}
		if number > float64(limit) {
			common.SysError(fmt.Sprintf("async task usage fact %q exceeds limit %d", key, limit))
			return int64(limit), true, true
		}
		value, clamp := common.QuotaRoundChecked(number)
		if clamp != nil || value < 0 {
			return 0, true, true
		}
		return int64(value), true, false
	}
	return 0, false, false
}

func taskCostUsageFromResult(result *relaycommon.TaskInfo) (model.CostUsage, []string) {
	usage := model.CostUsage{IncludeRequestFee: false}
	if result == nil {
		return usage, nil
	}
	rejected := make([]string, 0)
	read := func(aliases ...string) int64 {
		value, present, invalid := readTaskUsageFact(result.UsageFacts, aliases, common.MaxQuota)
		if invalid {
			rejected = append(rejected, aliases[0])
		}
		if !present || invalid {
			return 0
		}
		usage.UsageAvailable = true
		return value
	}
	usage.PromptTokens = read("prompt_tokens", "promptTokens", "input_tokens", "inputTokens")
	usage.CompletionTokens = read("completion_tokens", "completionTokens", "output_tokens", "outputTokens")
	totalTokens := read("total_tokens", "totalTokens")
	if usage.CompletionTokens == 0 && result.CompletionTokens > 0 {
		usage.CompletionTokens = int64(result.CompletionTokens)
		usage.UsageAvailable = true
	}
	if totalTokens == 0 && result.TotalTokens > 0 {
		totalTokens = int64(result.TotalTokens)
		usage.UsageAvailable = true
	}
	if usage.PromptTokens == 0 && usage.CompletionTokens == 0 && totalTokens > 0 {
		usage.CompletionTokens = totalTokens
	} else if usage.PromptTokens == 0 && totalTokens >= usage.CompletionTokens {
		usage.PromptTokens = totalTokens - usage.CompletionTokens
	}
	usage.CachedTokens = read("cached_tokens", "cachedTokens", "cache_read_tokens", "cacheReadTokens")
	usage.CacheCreationTokens = read("cache_creation_tokens", "cacheCreationTokens", "cache_write_tokens", "cacheWriteTokens")
	usage.ImageTokens = read("image_tokens", "imageTokens")
	usage.AudioInputTokens = read("audio_input_tokens", "audioInputTokens")
	usage.AudioOutputTokens = read("audio_output_tokens", "audioOutputTokens")
	usage.ImageCount = read("image_count", "imageCount", "images", "count")
	readDuration := func(aliases ...string) int64 {
		value, present, invalid := readTaskUsageFact(result.UsageFacts, aliases, relaycommon.MaxTaskDurationSeconds)
		if invalid {
			rejected = append(rejected, aliases[0])
		}
		if !present || invalid {
			return 0
		}
		usage.UsageAvailable = true
		return value
	}
	usage.VideoSeconds = readDuration("video_seconds", "videoSeconds", "video_duration", "videoDuration", "video")
	usage.AudioSeconds = readDuration("audio_seconds", "audioSeconds", "audio_duration", "audioDuration", "audio", "audio_input_seconds", "audioInputSeconds")
	usage.AudioOutputSeconds = readDuration("audio_output_seconds", "audioOutputSeconds", "audio_output_duration", "audioOutputDuration")
	if usage.VideoSeconds == 0 {
		usage.VideoSeconds = readDuration("seconds", "duration")
	}
	return usage, rejected
}

func taskBillingOtherWithCostUsage(task *model.Task, usage model.CostUsage, event string, rejected []string) map[string]interface{} {
	other := taskBillingOther(task)
	other["cost_usage"] = map[string]interface{}{
		"usage_available": usage.UsageAvailable,
		"prompt_tokens": usage.PromptTokens,
		"completion_tokens": usage.CompletionTokens,
		"cached_tokens": usage.CachedTokens,
		"cache_creation_tokens": usage.CacheCreationTokens,
		"image_tokens": usage.ImageTokens,
		"audio_input_tokens": usage.AudioInputTokens,
		"audio_output_tokens": usage.AudioOutputTokens,
		"image_count": usage.ImageCount,
		"audio_seconds": usage.AudioSeconds,
		"audio_output_seconds": usage.AudioOutputSeconds,
		"video_seconds": usage.VideoSeconds,
	}
	other["cost_usage_event"] = event
	if len(rejected) > 0 {
		other["cost_usage_rejected"] = rejected
	}
	return other
}

// taskBillingOther 从 task 的 BillingContext 构建日志 Other 字段。
func taskBillingOther(task *model.Task) map[string]interface{} {
	other := make(map[string]interface{})
	if bc := task.PrivateData.BillingContext; bc != nil {
		other["model_price"] = bc.ModelPrice
		if bc.ModelRatio > 0 {
			other["model_ratio"] = bc.ModelRatio
		}
		other["group_ratio"] = bc.GroupRatio
		if bc.BillingMode != "" {
			other["billing_mode"] = bc.BillingMode
		}
		if bc.BillingResolution != "" {
			other["resolution"] = bc.BillingResolution
		}
		if priceData := taskBillingContextPriceData(bc); priceData != nil {
			for k, v := range priceData.OtherRatios() {
				other[k] = v
			}
		}
	}
	return other
}

func taskBillingContextPriceData(bc *model.TaskBillingContext) *types.PriceData {
	if bc == nil || len(bc.OtherRatios) == 0 {
		return nil
	}
	priceData := &types.PriceData{}
	if !priceData.ReplaceOtherRatios(bc.OtherRatios) {
		return nil
	}
	return priceData
}

// taskModelName 从 BillingContext 或 Properties 中获取模型名称。
func taskModelName(task *model.Task) string {
	if bc := task.PrivateData.BillingContext; bc != nil && bc.OriginModelName != "" {
		return bc.OriginModelName
	}
	return task.Properties.OriginModelName
}

// RefundTaskQuota 统一的任务失败退款逻辑。
// 当异步任务失败时，将预扣的 quota 退还给用户（支持钱包和订阅），并退还令牌额度。
// 返回资金来源是否已成功退还；失败时保留 quota，供显式重试或人工对账。
func RefundTaskQuota(ctx context.Context, task *model.Task, reason string) bool {
	quota := task.Quota
	if quota == 0 {
		return true
	}

	// 1. 退还资金来源（钱包或订阅）
	if err := taskAdjustFunding(task, -quota); err != nil {
		logger.LogWarn(ctx, fmt.Sprintf("退还资金来源失败 task %s: %s", task.TaskID, err.Error()))
		return false
	}

	// 2. 退还令牌额度
	taskAdjustTokenQuota(ctx, task, -quota)

	// 3. 记录日志
	other := taskBillingOther(task)
	other["task_id"] = task.TaskID
	other["reason"] = reason
	model.RecordTaskBillingLog(model.RecordTaskBillingLogParams{
		UserId:    task.UserId,
		LogType:   model.LogTypeRefund,
		Content:   "",
		ChannelId: task.ChannelId,
		ModelName: taskModelName(task),
		Quota:     quota,
		TokenId:   task.PrivateData.TokenId,
		Group:     task.Group,
		Other:     other,
	})

	// 4. 资金退款完成后再清除持久化标记。
	// 回写失败必须显式告警，避免漏掉潜在的重复退款风险。
	task.Quota = 0
	if err := task.UpdateQuota(); err != nil {
		logger.LogError(ctx, fmt.Sprintf("退款成功但清除 task quota 失败 task %s: %s", task.TaskID, err.Error()))
	}
	return true
}

// RecalculateTaskQuota 通用的异步差额结算。
// actualQuota 是任务完成后的实际应扣额度，与预扣额度 (task.Quota) 做差额结算。
// reason 用于日志记录（例如 "token重算" 或 "adaptor调整"）。
// clamps 可选：若计算 actualQuota 时发生额度饱和，将其记入日志 admin_info（仅管理员可见）。
func RecalculateTaskQuota(ctx context.Context, task *model.Task, actualQuota int, reason string, clamps ...*common.QuotaClamp) {
	recalculateTaskQuotaWithUsage(ctx, task, actualQuota, reason, model.CostUsage{}, nil, clamps...)
}

func recalculateTaskQuotaWithUsage(ctx context.Context, task *model.Task, actualQuota int, reason string, usage model.CostUsage, rejected []string, clamps ...*common.QuotaClamp) {
	if actualQuota <= 0 || task == nil {
		return
	}
	preConsumedQuota := task.Quota
	quotaDelta := actualQuota - preConsumedQuota

	if quotaDelta == 0 {
		logger.LogInfo(ctx, fmt.Sprintf("任务 %s 预扣费准确（%s，%s）",
			task.TaskID, logger.LogQuota(actualQuota), reason))
		if usage.UsageAvailable || len(rejected) > 0 {
			other := taskBillingOtherWithCostUsage(task, usage, "settlement", rejected)
			other["task_id"] = task.TaskID
			other["pre_consumed_quota"] = preConsumedQuota
			other["actual_quota"] = actualQuota
			model.RecordTaskBillingLog(model.RecordTaskBillingLogParams{UserId: task.UserId, LogType: model.LogTypeConsume, Content: reason, ChannelId: task.ChannelId, ModelName: taskModelName(task), TokenId: task.PrivateData.TokenId, Group: task.Group, Other: other, CostUsage: &usage})
		}
		return
	}

	logger.LogInfo(ctx, fmt.Sprintf("任务 %s 差额结算：delta=%s（实际：%s，预扣：%s，%s）",
		task.TaskID,
		logger.LogQuota(quotaDelta),
		logger.LogQuota(actualQuota),
		logger.LogQuota(preConsumedQuota),
		reason,
	))

	// 调整资金来源
	if err := taskAdjustFunding(task, quotaDelta); err != nil {
		logger.LogError(ctx, fmt.Sprintf("差额结算资金调整失败 task %s: %s", task.TaskID, err.Error()))
		return
	}

	// 调整令牌额度
	taskAdjustTokenQuota(ctx, task, quotaDelta)

	task.Quota = actualQuota
	if err := task.UpdateQuota(); err != nil {
		logger.LogError(ctx, fmt.Sprintf("差额结算回写 quota 失败 task %s: %s", task.TaskID, err.Error()))
	}

	var logType int
	var logQuota int
	if quotaDelta > 0 {
		logType = model.LogTypeConsume
		logQuota = quotaDelta
		model.UpdateUserUsedQuotaAndRequestCount(task.UserId, quotaDelta)
		model.UpdateChannelUsedQuota(task.ChannelId, quotaDelta)
	} else {
		logType = model.LogTypeRefund
		logQuota = -quotaDelta
	}
	if quotaDelta < 0 {
		usage = model.CostUsage{}
		rejected = nil
	}
	other := taskBillingOtherWithCostUsage(task, usage, map[bool]string{true: "settlement", false: "refund"}[quotaDelta > 0], rejected)
	other["task_id"] = task.TaskID
	other["pre_consumed_quota"] = preConsumedQuota
	other["actual_quota"] = actualQuota
	for _, clamp := range clamps {
		attachQuotaSaturationToOther(other, clamp)
	}
	model.RecordTaskBillingLog(model.RecordTaskBillingLogParams{
		UserId:    task.UserId,
		LogType:   logType,
		Content:   reason,
		ChannelId: task.ChannelId,
		ModelName: taskModelName(task),
		Quota:     logQuota,
		TokenId:   task.PrivateData.TokenId,
		Group:     task.Group,
		Other:     other,
		NodeName:  task.PrivateData.NodeName,
		CostUsage: func() *model.CostUsage { if logType == model.LogTypeConsume { return &usage }; return nil }(),
	})
}

// RecalculateTaskQuotaByTokens 根据实际 token 消耗重新计费（异步差额结算）。
// 当任务成功且返回了 totalTokens 时，根据模型倍率和分组倍率重新计算实际扣费额度，
// 与预扣费的差额进行补扣或退还。支持钱包和订阅计费来源。
func RecalculateTaskQuotaByTokens(ctx context.Context, task *model.Task, totalTokens int) {
	if totalTokens <= 0 {
		return
	}

	modelName := taskModelName(task)

	// 获取模型价格和倍率
	modelRatio, hasRatioSetting, _ := ratio_setting.GetModelRatio(modelName)
	// 只有配置了倍率(非固定价格)时才按 token 重新计费
	if !hasRatioSetting || modelRatio <= 0 {
		return
	}

	// 获取用户和组的倍率信息
	group := task.Group
	if group == "" {
		user, err := model.GetUserById(task.UserId, false)
		if err == nil {
			group = user.Group
		}
	}
	if group == "" {
		return
	}

	// New tasks persist the exact effective group ratio used during pre-consume.
	// Use that snapshot for settlement, including an explicit zero ratio. Only
	// legacy tasks without a billing context should resolve the current settings.
	finalGroupRatio := 0.0
	if billingContext := task.PrivateData.BillingContext; billingContext != nil {
		finalGroupRatio = billingContext.GroupRatio
	} else {
		groupRatio := ratio_setting.GetGroupRatio(group)
		userGroupRatio, hasUserGroupRatio := ratio_setting.GetGroupGroupRatio(group, group)
		if hasUserGroupRatio {
			finalGroupRatio = userGroupRatio
		} else {
			finalGroupRatio = groupRatio
		}
	}

	// 计算 OtherRatios 乘积（视频折扣、时长等）
	otherMultiplier := 1.0
	if priceData := taskBillingContextPriceData(task.PrivateData.BillingContext); priceData != nil {
		otherMultiplier = priceData.OtherRatioMultiplier()
	}

	// 计算实际应扣费额度: totalTokens * modelRatio * groupRatio * otherMultiplier（饱和转换，防止溢出成负数）
	actualQuota, clamp := common.QuotaFromFloatChecked(float64(totalTokens) * modelRatio * finalGroupRatio * otherMultiplier * common.QuotaPerUnit)

	reason := fmt.Sprintf("token重算：tokens=%d, modelRatio=%.2f, groupRatio=%.2f, otherMultiplier=%.4f", totalTokens, modelRatio, finalGroupRatio, otherMultiplier)
	usage := model.CostUsage{UsageAvailable: true, CompletionTokens: int64(totalTokens)}
	recalculateTaskQuotaWithUsage(ctx, task, actualQuota, reason, usage, nil, clamp)
}
