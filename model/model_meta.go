package model

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"

	"gorm.io/gorm"
)

const (
	NameRuleExact = iota
	NameRulePrefix
	NameRuleContains
	NameRuleSuffix
)

type BoundChannel struct {
	Id            int    `json:"id"`
	Name          string `json:"name"`
	Type          int    `json:"type"`
	Status        int    `json:"status"`
	UpstreamModel string `json:"upstream_model"`
}

type Model struct {
	Id           int            `json:"id"`
	ModelName    string         `json:"model_name" gorm:"size:128;not null;uniqueIndex:uk_model_name_delete_at,priority:1"`
	ModelType    string         `json:"model_type" gorm:"size:32;index"`
	Description  string         `json:"description,omitempty" gorm:"type:text"`
	Icon         string         `json:"icon,omitempty" gorm:"type:varchar(128)"`
	Tags         string         `json:"tags,omitempty" gorm:"type:varchar(255)"`
	VendorID     int            `json:"vendor_id,omitempty" gorm:"index"`
	Endpoints    string         `json:"endpoints,omitempty" gorm:"type:text"`
	Status       int            `json:"status" gorm:"default:1"`
	SyncOfficial int            `json:"sync_official" gorm:"default:1"`
	CreatedTime  int64          `json:"created_time" gorm:"bigint"`
	UpdatedTime  int64          `json:"updated_time" gorm:"bigint"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index;uniqueIndex:uk_model_name_delete_at,priority:2"`

	BoundChannels []BoundChannel `json:"bound_channels,omitempty" gorm:"-"`
	EnableGroups  []string       `json:"enable_groups,omitempty" gorm:"-"`
	QuotaTypes    []int          `json:"quota_types,omitempty" gorm:"-"`
	NameRule      int            `json:"name_rule" gorm:"default:0"`

	MatchedModels []string `json:"matched_models,omitempty" gorm:"-"`
	MatchedCount  int      `json:"matched_count,omitempty" gorm:"-"`
}

func (mi *Model) Insert() error {
	now := common.GetTimestamp()
	mi.CreatedTime = now
	mi.UpdatedTime = now

	// 保存原始值（因为 Create 后可能被 GORM 的 default 标签覆盖为 1）
	originalStatus := mi.Status
	originalSyncOfficial := mi.SyncOfficial

	// 先创建记录（GORM 会对零值字段应用默认值）
	if err := DB.Create(mi).Error; err != nil {
		return err
	}

	// 使用保存的原始值进行更新，确保零值能正确保存
	return DB.Model(&Model{}).Where("id = ?", mi.Id).Updates(map[string]interface{}{
		"status":        originalStatus,
		"sync_official": originalSyncOfficial,
	}).Error
}

func IsModelNameDuplicated(id int, name string) (bool, error) {
	if name == "" {
		return false, nil
	}
	var cnt int64
	err := DB.Model(&Model{}).Where("model_name = ? AND id <> ?", name, id).Count(&cnt).Error
	return cnt > 0, err
}

func (mi *Model) Update() error {
	mi.UpdatedTime = common.GetTimestamp()
	// 使用 Select 强制更新所有字段，包括零值
	return DB.Model(&Model{}).Where("id = ?", mi.Id).
		Select("model_name", "model_type", "description", "icon", "tags", "vendor_id", "endpoints", "status", "sync_official", "name_rule", "updated_time").
		Updates(mi).Error
}

func (mi *Model) Delete() error {
	return DB.Delete(mi).Error
}

func GetVendorModelCounts() (map[int64]int64, error) {
	var stats []struct {
		VendorID int64
		Count    int64
	}
	if err := DB.Model(&Model{}).
		Select("vendor_id as vendor_id, count(*) as count").
		Group("vendor_id").
		Scan(&stats).Error; err != nil {
		return nil, err
	}
	m := make(map[int64]int64, len(stats))
	for _, s := range stats {
		m[s.VendorID] = s.Count
	}
	return m, nil
}

func GetAllModels(offset int, limit int) ([]*Model, error) {
	models, _, err := SearchModels("", "", "", "", 0, offset, limit)
	return models, err
}

func GetBoundChannelsByModelsMap(modelNames []string) (map[string][]BoundChannel, error) {
	result := make(map[string][]BoundChannel)
	if len(modelNames) == 0 {
		return result, nil
	}
	type row struct {
		ID           int
		Model        string
		Name         string
		Type         int
		Status       int
		ModelMapping *string
	}
	// Disabled channels are included (with their status) so the admin UI shows
	// every binding and manual re-binding never drops a disabled channel.
	var rows []row
	err := DB.Table("channels").
		Select("channels.id as id, abilities.model as model, channels.name as name, channels.type as type, channels.status as status, channels.model_mapping as model_mapping").
		Joins("JOIN abilities ON abilities.channel_id = channels.id").
		Where("abilities.model IN ?", modelNames).
		Distinct().
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		upstreamModel := r.Model
		if r.ModelMapping != nil {
			if mapped, _, mappingErr := common.ResolveModelMapping(*r.ModelMapping, r.Model); mappingErr == nil {
				upstreamModel = mapped
			}
		}
		result[r.Model] = append(result[r.Model], BoundChannel{
			Id:            r.ID,
			Name:          r.Name,
			Type:          r.Type,
			Status:        r.Status,
			UpstreamModel: upstreamModel,
		})
	}
	return result, nil
}

func normalizeLookupValues(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}
	return normalized
}

func GetPreferredModelOwnerChannelTypes(modelNames []string, groups []string) (map[string]int, error) {
	result := make(map[string]int)
	modelNames = normalizeLookupValues(modelNames)
	if len(modelNames) == 0 {
		return result, nil
	}

	type row struct {
		Model       string
		ChannelType int
	}
	var rows []row

	query := DB.Table("abilities").
		Select("abilities.model as model, channels.type as channel_type").
		Joins("JOIN channels ON abilities.channel_id = channels.id").
		Where("abilities.model IN ? AND abilities.enabled = ? AND channels.status = ?", modelNames, true, common.ChannelStatusEnabled).
		Order("COALESCE(abilities.priority, 0) DESC").
		Order("abilities.weight DESC").
		Order("abilities.channel_id ASC")

	groups = normalizeLookupValues(groups)
	if len(groups) > 0 {
		query = query.Where("abilities."+commonGroupCol+" IN ?", groups)
	}

	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}

	for _, r := range rows {
		if _, ok := result[r.Model]; ok {
			continue
		}
		result[r.Model] = r.ChannelType
	}
	return result, nil
}

func SearchModels(keyword string, vendor string, status string, syncOfficial string, channelID int, offset int, limit int) ([]*Model, int64, error) {
	var models []*Model
	db := DB.Model(&Model{})
	if keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("model_name LIKE ? OR description LIKE ? OR tags LIKE ?", like, like, like)
	}
	if vendor != "" {
		if vid, err := strconv.Atoi(vendor); err == nil {
			db = db.Where("models.vendor_id = ?", vid)
		} else {
			db = db.Joins("JOIN vendors ON vendors.id = models.vendor_id").Where("vendors.name LIKE ?", "%"+vendor+"%")
		}
	}
	if statusValue, ok := parseModelStatusFilter(status); ok {
		db = db.Where("models.status = ?", statusValue)
	}
	if syncValue, ok := parseModelSyncFilter(syncOfficial); ok {
		db = db.Where("models.sync_official = ?", syncValue)
	}
	if channelID > 0 {
		db = db.Where("EXISTS (SELECT 1 FROM abilities WHERE abilities.model = models.model_name AND abilities.channel_id = ? AND abilities.enabled = ?)", channelID, true)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Order("models.id DESC").Offset(offset).Limit(limit).Find(&models).Error; err != nil {
		return nil, 0, err
	}
	return models, total, nil
}

// UpdateModelChannelBindings synchronizes a model's explicit channel selection
// with the channel model lists and their derived abilities. Automatic binding
// intentionally keeps the existing channel configuration unchanged.
func UpdateModelChannelBindings(modelID int, mode string, channelIDs []int) error {
	if mode != "manual" {
		return nil
	}

	selected := make(map[int]struct{}, len(channelIDs))
	for _, channelID := range channelIDs {
		if channelID > 0 {
			selected[channelID] = struct{}{}
		}
	}

	err := DB.Transaction(func(tx *gorm.DB) error {
		var target Model
		if err := lockForUpdate(tx).First(&target, modelID).Error; err != nil {
			return err
		}

		var boundIDs []int
		if err := tx.Model(&Ability{}).
			Where("model = ?", target.ModelName).
			Distinct("channel_id").
			Pluck("channel_id", &boundIDs).Error; err != nil {
			return err
		}
		allIDs := make(map[int]struct{}, len(boundIDs)+len(selected))
		for _, channelID := range boundIDs {
			allIDs[channelID] = struct{}{}
		}
		for channelID := range selected {
			allIDs[channelID] = struct{}{}
		}
		if len(allIDs) == 0 {
			return nil
		}

		ids := make([]int, 0, len(allIDs))
		for channelID := range allIDs {
			ids = append(ids, channelID)
		}
		var channels []Channel
		if err := lockForUpdate(tx).Where("id IN ?", ids).Find(&channels).Error; err != nil {
			return err
		}
		if len(channels) != len(ids) {
			return fmt.Errorf("one or more channels do not exist")
		}

		for i := range channels {
			channel := &channels[i]
			_, shouldBind := selected[channel.Id]
			if shouldBind {
				channel.Models = appendModelName(channel.Models, target.ModelName)
			} else {
				channel.Models = removeModelName(channel.Models, target.ModelName)
				mapping, err := removeModelMapping(channel.GetModelMapping(), target.ModelName)
				if err != nil {
					return err
				}
				channel.ModelMapping = &mapping
			}
			if err := tx.Model(&Channel{}).Where("id = ?", channel.Id).
				Select("models", "model_mapping").Updates(channel).Error; err != nil {
				return err
			}
			if err := channel.UpdateAbilities(tx); err != nil {
				return err
			}
		}
		return nil
	})
	if err == nil {
		InitChannelCache()
	}
	return err
}

func appendModelName(value, modelName string) string {
	parts := strings.Split(value, ",")
	seen := make(map[string]struct{}, len(parts)+1)
	result := make([]string, 0, len(parts)+1)
	for _, part := range append(parts, modelName) {
		name := strings.TrimSpace(part)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		result = append(result, name)
	}
	return strings.Join(result, ",")
}

func removeModelName(value, modelName string) string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		name := strings.TrimSpace(part)
		if name != "" && name != modelName {
			result = append(result, name)
		}
	}
	return strings.Join(result, ",")
}

func removeModelMapping(value, modelName string) (string, error) {
	mapping := make(map[string]string)
	if strings.TrimSpace(value) != "" && strings.TrimSpace(value) != "{}" {
		if err := common.UnmarshalJsonStr(value, &mapping); err != nil {
			return "", fmt.Errorf("invalid channel model mapping: %w", err)
		}
	}
	delete(mapping, modelName)
	data, err := common.Marshal(mapping)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// parseModelStatusFilter maps UI/API status values to the models.status column.
// Returns ok=false when no status filter should be applied.
func parseModelStatusFilter(status string) (value int, ok bool) {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "", "all":
		return 0, false
	case "enabled", "1":
		return 1, true
	case "disabled", "0":
		return 0, true
	default:
		n, err := strconv.Atoi(status)
		if err != nil {
			return 0, false
		}
		return n, true
	}
}

// parseModelSyncFilter maps UI/API sync values to the models.sync_official column.
// Returns ok=false when no sync filter should be applied.
func parseModelSyncFilter(syncOfficial string) (value int, ok bool) {
	switch strings.ToLower(strings.TrimSpace(syncOfficial)) {
	case "", "all":
		return 0, false
	case "yes", "1":
		return 1, true
	case "no", "0":
		return 0, true
	default:
		n, err := strconv.Atoi(syncOfficial)
		if err != nil {
			return 0, false
		}
		return n, true
	}
}
