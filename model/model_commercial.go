package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

var modelKeyedOptionKeys = []string{
	"ModelPrice",
	"ModelRatio",
	"CacheRatio",
	"CreateCacheRatio",
	"CompletionRatio",
	"ImageRatio",
	"ImageResolutionPrice",
	"AudioRatio",
	"AudioCompletionRatio",
	"billing_setting.billing_mode",
	"billing_setting.billing_expr",
	"billing_setting.task_billing_pricing",
	"billing_setting.scheduled_discount",
	"claude.model_headers_settings",
	"claude.default_max_tokens",
	"gemini.safety_settings",
	"gemini.version_settings",
}

func updateModelNameInList(value, oldName, newName string) string {
	parts := strings.Split(value, ",")
	seen := make(map[string]struct{}, len(parts))
	updated := make([]string, 0, len(parts))
	for _, part := range parts {
		name := strings.TrimSpace(part)
		if name == oldName {
			name = newName
		}
		if name == "" {
			continue
		}
		if _, exists := seen[name]; exists {
			continue
		}
		seen[name] = struct{}{}
		updated = append(updated, name)
	}
	return strings.Join(updated, ",")
}

func updateChannelModelMapping(value, oldName, newName, upstreamModel string) (string, error) {
	mapping := make(map[string]string)
	if strings.TrimSpace(value) != "" && strings.TrimSpace(value) != "{}" {
		if err := common.UnmarshalJsonStr(value, &mapping); err != nil {
			return "", fmt.Errorf("invalid channel model mapping: %w", err)
		}
	}

	if oldName != newName {
		if mapped, exists := mapping[oldName]; exists {
			mapping[newName] = mapped
			delete(mapping, oldName)
		}
	}
	if upstreamModel != "" {
		if upstreamModel == newName {
			delete(mapping, newName)
		} else {
			mapping[newName] = upstreamModel
		}
	}

	data, err := common.Marshal(mapping)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func saveOptionWithTx(tx *gorm.DB, key, value string) error {
	option := Option{}
	err := lockForUpdate(tx).Where("key = ?", key).First(&option).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return tx.Create(&Option{Key: key, Value: value}).Error
	}
	if err != nil {
		return err
	}
	return tx.Model(&Option{}).Where("key = ?", key).Update("value", value).Error
}

func migrateModelOptionKeys(tx *gorm.DB, oldName, newName string) (map[string]string, error) {
	keys := append([]string(nil), modelKeyedOptionKeys...)
	keys = append(keys, modelCostOptionKey)
	var options []Option
	if err := lockForUpdate(tx).Where("key IN ?", keys).Find(&options).Error; err != nil {
		return nil, err
	}

	updates := make(map[string]string)
	for _, option := range options {
		values := make(map[string]json.RawMessage)
		if strings.TrimSpace(option.Value) == "" {
			continue
		}
		if err := common.UnmarshalJsonStr(option.Value, &values); err != nil {
			return nil, fmt.Errorf("invalid %s option: %w", option.Key, err)
		}
		changed := false
		if value, exists := values[oldName]; exists {
			values[newName] = value
			delete(values, oldName)
			changed = true
		}
		if option.Key == modelCostOptionKey {
			for key, value := range values {
				if !strings.HasPrefix(key, "channel:") {
					continue
				}
				parts := strings.SplitN(strings.TrimPrefix(key, "channel:"), ":", 2)
				if len(parts) != 2 || parts[1] != oldName {
					continue
				}
				newKey := "channel:" + parts[0] + ":" + newName
				values[newKey] = value
				delete(values, key)
				changed = true
			}
		}
		if !changed {
			continue
		}
		data, err := common.Marshal(values)
		if err != nil {
			return nil, err
		}
		updates[option.Key] = string(data)
		if err := tx.Model(&Option{}).Where("key = ?", option.Key).Update("value", updates[option.Key]).Error; err != nil {
			return nil, err
		}
	}
	return updates, nil
}

func (mi *Model) UpdateWithConfigurationMigration() error {
	optionUpdates := make(map[string]string)
	channelsChanged := false
	err := DB.Transaction(func(tx *gorm.DB) error {
		var existing Model
		if err := lockForUpdate(tx).First(&existing, mi.Id).Error; err != nil {
			return err
		}

		mi.UpdatedTime = common.GetTimestamp()
		if err := tx.Model(&Model{}).Where("id = ?", mi.Id).
			Select("model_name", "model_type", "description", "icon", "tags", "vendor_id", "endpoints", "status", "sync_official", "name_rule", "updated_time").
			Updates(mi).Error; err != nil {
			return err
		}
		if existing.ModelName == mi.ModelName {
			return nil
		}

		var channelIDs []int
		if err := tx.Model(&Ability{}).
			Where("model = ?", existing.ModelName).
			Distinct("channel_id").
			Pluck("channel_id", &channelIDs).Error; err != nil {
			return err
		}
		if len(channelIDs) > 0 {
			var channels []Channel
			if err := lockForUpdate(tx).Where("id IN ?", channelIDs).Find(&channels).Error; err != nil {
				return err
			}
			for i := range channels {
				channel := &channels[i]
				mapping, err := updateChannelModelMapping(channel.GetModelMapping(), existing.ModelName, mi.ModelName, "")
				if err != nil {
					return fmt.Errorf("channel %d: %w", channel.Id, err)
				}
				channel.Models = updateModelNameInList(channel.Models, existing.ModelName, mi.ModelName)
				channel.ModelMapping = &mapping
				if err := tx.Model(&Channel{}).Where("id = ?", channel.Id).
					Select("models", "model_mapping").Updates(channel).Error; err != nil {
					return err
				}
				if err := channel.UpdateAbilities(tx); err != nil {
					return err
				}
			}
			channelsChanged = true
		}

		var err error
		optionUpdates, err = migrateModelOptionKeys(tx, existing.ModelName, mi.ModelName)
		return err
	})
	if err != nil {
		return err
	}
	for key, value := range optionUpdates {
		if err := updateOptionMap(key, value); err != nil {
			common.SysError(fmt.Sprintf("refresh renamed model option %s: %v", key, err))
		}
	}
	if channelsChanged {
		InitChannelCache()
	}
	return nil
}

func SaveModelCommercialConfig(modelID, channelID int, upstreamModel string, cost ModelCost) error {
	var optionValue string
	channelChanged := false
	err := DB.Transaction(func(tx *gorm.DB) error {
		var targetModel Model
		if err := lockForUpdate(tx).First(&targetModel, modelID).Error; err != nil {
			return err
		}

		costKey := targetModel.ModelName
		if channelID > 0 {
			var abilityCount int64
			if err := tx.Model(&Ability{}).
				Where("channel_id = ? AND model = ?", channelID, targetModel.ModelName).
				Count(&abilityCount).Error; err != nil {
				return err
			}
			if abilityCount == 0 {
				return fmt.Errorf("channel %d is not bound to model %s", channelID, targetModel.ModelName)
			}

			var channel Channel
			if err := lockForUpdate(tx).First(&channel, channelID).Error; err != nil {
				return err
			}
			upstreamModel = strings.TrimSpace(upstreamModel)
			if upstreamModel == "" {
				upstreamModel = targetModel.ModelName
			}
			if len(upstreamModel) > 255 {
				return fmt.Errorf("upstream model name must be at most 255 characters")
			}
			mapping, err := updateChannelModelMapping(channel.GetModelMapping(), targetModel.ModelName, targetModel.ModelName, upstreamModel)
			if err != nil {
				return err
			}
			if err := tx.Model(&Channel{}).Where("id = ?", channelID).Update("model_mapping", mapping).Error; err != nil {
				return err
			}
			channelChanged = true
			costKey = fmt.Sprintf("channel:%d:%s", channelID, targetModel.ModelName)
		}

		normalizedCost, err := validatedModelCost(costKey, cost)
		if err != nil {
			return err
		}
		option := Option{}
		queryErr := lockForUpdate(tx).Where("key = ?", modelCostOptionKey).First(&option).Error
		if queryErr != nil && !errors.Is(queryErr, gorm.ErrRecordNotFound) {
			return queryErr
		}
		costs := make(ModelCostConfig)
		if queryErr == nil && strings.TrimSpace(option.Value) != "" {
			if err := common.UnmarshalJsonStr(option.Value, &costs); err != nil {
				return fmt.Errorf("invalid ModelCost option: %w", err)
			}
		}
		costs[costKey] = normalizedCost
		data, err := common.Marshal(costs)
		if err != nil {
			return err
		}
		optionValue = string(data)
		return saveOptionWithTx(tx, modelCostOptionKey, optionValue)
	})
	if err != nil {
		return err
	}
	if err := updateOptionMap(modelCostOptionKey, optionValue); err != nil {
		common.SysError("refresh ModelCost option: " + err.Error())
	}
	if channelChanged {
		InitChannelCache()
	}
	return nil
}
