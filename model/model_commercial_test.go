package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func prepareModelCommercialTest(t *testing.T) {
	t.Helper()
	require.NoError(t, DB.AutoMigrate(&Model{}, &Option{}))
	require.NoError(t, DB.Where("id >= ?", 910000).Delete(&Channel{}).Error)
	require.NoError(t, DB.Where("id >= ?", 910000).Delete(&Model{}).Error)
	require.NoError(t, DB.Where("channel_id >= ?", 910000).Delete(&Ability{}).Error)
	require.NoError(t, DB.Where("key IN ?", append(append([]string(nil), modelKeyedOptionKeys...), modelCostOptionKey)).Delete(&Option{}).Error)

	var previousOptionMap map[string]string
	previousOptions := make(map[string]string)
	common.OptionMapRWMutex.Lock()
	previousOptionMap = common.OptionMap
	if common.OptionMap == nil {
		common.OptionMap = make(map[string]string)
	}
	for _, key := range append(append([]string(nil), modelKeyedOptionKeys...), modelCostOptionKey) {
		previousOptions[key] = common.OptionMap[key]
	}
	common.OptionMapRWMutex.Unlock()
	t.Cleanup(func() {
		_ = DB.Where("channel_id >= ?", 910000).Delete(&Ability{}).Error
		_ = DB.Where("id >= ?", 910000).Delete(&Channel{}).Error
		_ = DB.Unscoped().Where("id >= ?", 910000).Delete(&Model{}).Error
		_ = DB.Where("key IN ?", append(append([]string(nil), modelKeyedOptionKeys...), modelCostOptionKey)).Delete(&Option{}).Error
		for key, value := range previousOptions {
			_ = updateOptionMap(key, value)
		}
		common.OptionMapRWMutex.Lock()
		common.OptionMap = previousOptionMap
		common.OptionMapRWMutex.Unlock()
		InitChannelCache()
	})
}

func createCommercialTestModelAndChannel(t *testing.T, modelName string, channelID int, upstreamModel string) Model {
	t.Helper()
	targetModel := Model{Id: channelID, ModelName: modelName, ModelType: "text", Status: 1, NameRule: NameRuleExact}
	require.NoError(t, DB.Create(&targetModel).Error)
	priority := int64(0)
	weight := uint(100)
	mapping, err := common.Marshal(map[string]string{modelName: upstreamModel})
	require.NoError(t, err)
	channel := Channel{
		Id:           channelID,
		Type:         1,
		Key:          "test-key",
		Status:       common.ChannelStatusEnabled,
		Name:         "test-channel",
		Models:       modelName + ",other-model",
		Group:        "default",
		Priority:     &priority,
		Weight:       &weight,
		ModelMapping: common.GetPointer(string(mapping)),
	}
	require.NoError(t, DB.Create(&channel).Error)
	require.NoError(t, channel.AddAbilities(nil))
	return targetModel
}

func TestSaveModelCommercialConfigUpdatesMappingAndScopedCostTogether(t *testing.T) {
	prepareModelCommercialTest(t)
	targetModel := createCommercialTestModelAndChannel(t, "public-model", 910001, "old-upstream")

	cost := ModelCost{Currency: "USD", InputPerMillion: 3, OutputPerMillion: 15, Enabled: true}
	require.NoError(t, SaveModelCommercialConfig(targetModel.Id, 910001, "new-upstream", cost))

	var channel Channel
	require.NoError(t, DB.First(&channel, 910001).Error)
	upstream, mapped, err := common.ResolveModelMapping(channel.GetModelMapping(), "public-model")
	require.NoError(t, err)
	assert.True(t, mapped)
	assert.Equal(t, "new-upstream", upstream)

	var option Option
	require.NoError(t, DB.Where("key = ?", modelCostOptionKey).First(&option).Error)
	var costs ModelCostConfig
	require.NoError(t, common.UnmarshalJsonStr(option.Value, &costs))
	assert.Equal(t, cost, costs["channel:910001:public-model"])

	channels, err := GetBoundChannelsByModelsMap([]string{"public-model"})
	require.NoError(t, err)
	require.Len(t, channels["public-model"], 1)
	assert.Equal(t, "new-upstream", channels["public-model"][0].UpstreamModel)
}

func TestUpdateWithConfigurationMigrationRenamesRoutingPricingAndCosts(t *testing.T) {
	prepareModelCommercialTest(t)
	targetModel := createCommercialTestModelAndChannel(t, "old-public-model", 910002, "provider-model")
	options := []Option{
		{Key: "ModelPrice", Value: `{"old-public-model":1.25,"other-model":2}`},
		{Key: "billing_setting.billing_mode", Value: `{"old-public-model":"per-request"}`},
		{Key: modelCostOptionKey, Value: `{"old-public-model":{"currency":"USD","enabled":true,"request_fee":0.5},"channel:910002:old-public-model":{"currency":"USD","enabled":true,"request_fee":0.4}}`},
	}
	require.NoError(t, DB.Create(&options).Error)

	targetModel.ModelName = "new-public-model"
	require.NoError(t, targetModel.UpdateWithConfigurationMigration())

	var channel Channel
	require.NoError(t, DB.First(&channel, 910002).Error)
	assert.Equal(t, "new-public-model,other-model", channel.Models)
	upstream, mapped, err := common.ResolveModelMapping(channel.GetModelMapping(), "new-public-model")
	require.NoError(t, err)
	assert.True(t, mapped)
	assert.Equal(t, "provider-model", upstream)

	var oldAbilityCount int64
	require.NoError(t, DB.Model(&Ability{}).Where("channel_id = ? AND model = ?", 910002, "old-public-model").Count(&oldAbilityCount).Error)
	assert.Zero(t, oldAbilityCount)
	var newAbilityCount int64
	require.NoError(t, DB.Model(&Ability{}).Where("channel_id = ? AND model = ?", 910002, "new-public-model").Count(&newAbilityCount).Error)
	assert.Equal(t, int64(1), newAbilityCount)

	var priceOption Option
	require.NoError(t, DB.Where("key = ?", "ModelPrice").First(&priceOption).Error)
	var prices map[string]float64
	require.NoError(t, common.UnmarshalJsonStr(priceOption.Value, &prices))
	assert.Equal(t, 1.25, prices["new-public-model"])
	assert.NotContains(t, prices, "old-public-model")

	var costOption Option
	require.NoError(t, DB.Where("key = ?", modelCostOptionKey).First(&costOption).Error)
	var costs ModelCostConfig
	require.NoError(t, common.UnmarshalJsonStr(costOption.Value, &costs))
	assert.Contains(t, costs, "new-public-model")
	assert.Contains(t, costs, "channel:910002:new-public-model")
	assert.NotContains(t, costs, "old-public-model")
	assert.NotContains(t, costs, "channel:910002:old-public-model")
}

func TestUpdateWithConfigurationMigrationRollsBackMalformedChannelMapping(t *testing.T) {
	prepareModelCommercialTest(t)
	targetModel := createCommercialTestModelAndChannel(t, "rollback-model", 910003, "provider-model")
	require.NoError(t, DB.Model(&Channel{}).Where("id = ?", 910003).Update("model_mapping", "{invalid").Error)

	targetModel.ModelName = "renamed-model"
	require.Error(t, targetModel.UpdateWithConfigurationMigration())

	var storedModel Model
	require.NoError(t, DB.First(&storedModel, targetModel.Id).Error)
	assert.Equal(t, "rollback-model", storedModel.ModelName)
	var channel Channel
	require.NoError(t, DB.First(&channel, 910003).Error)
	assert.Equal(t, "rollback-model,other-model", channel.Models)
}
