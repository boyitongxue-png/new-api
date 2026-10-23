package model

import (
	"fmt"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupModelMetaTestDB(t *testing.T) {
	t.Helper()
	previousDB := DB
	previousMemoryCacheEnabled := common.MemoryCacheEnabled
	previousMainDatabaseType := common.MainDatabaseType()
	previousLogDatabaseType := common.LogDatabaseType()
	common.MemoryCacheEnabled = false
	common.SetDatabaseTypes(common.DatabaseTypeSQLite, common.DatabaseTypeSQLite)
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	require.NoError(t, err)
	DB = db
	require.NoError(t, DB.AutoMigrate(&Model{}, &Channel{}, &Ability{}, &Vendor{}))
	t.Cleanup(func() {
		DB = previousDB
		common.MemoryCacheEnabled = previousMemoryCacheEnabled
		common.SetDatabaseTypes(previousMainDatabaseType, previousLogDatabaseType)
	})
}

func TestSearchModelsFiltersByEnabledChannelAbility(t *testing.T) {
	setupModelMetaTestDB(t)
	models := []Model{{ModelName: "bound-model"}, {ModelName: "other-model"}}
	require.NoError(t, DB.Create(&models).Error)
	require.NoError(t, DB.Create(&Channel{Id: 101, Name: "channel-a", Key: "key-a", Models: "bound-model"}).Error)
	require.NoError(t, DB.Create(&Channel{Id: 102, Name: "channel-b", Key: "key-b", Models: "other-model"}).Error)
	require.NoError(t, DB.Create(&[]Ability{
		{Group: "default", Model: "bound-model", ChannelId: 101, Enabled: true},
		{Group: "default", Model: "other-model", ChannelId: 102, Enabled: false},
	}).Error)

	got, total, err := SearchModels("", "", "", "", 101, 0, 10)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, got, 1)
	assert.Equal(t, "bound-model", got[0].ModelName)
}

func TestUpdateModelChannelBindingsManualUpdatesChannelsAndAbilities(t *testing.T) {
	setupModelMetaTestDB(t)
	modelRow := Model{ModelName: "manual-model"}
	require.NoError(t, DB.Create(&modelRow).Error)
	channels := []Channel{
		{Id: 201, Name: "selected", Key: "key-201", Models: "existing-model"},
		{Id: 202, Name: "removed", Key: "key-202", Models: "manual-model,other-model", ModelMapping: strptr(`{"manual-model":"upstream-model","other-model":"other-upstream"}`)},
	}
	require.NoError(t, DB.Create(&channels).Error)
	require.NoError(t, DB.Create(&[]Ability{
		{Group: "default", Model: "manual-model", ChannelId: 202, Enabled: true},
		{Group: "default", Model: "other-model", ChannelId: 202, Enabled: true},
	}).Error)

	require.NoError(t, UpdateModelChannelBindings(modelRow.Id, "manual", []int{201}))

	var selected Channel
	require.NoError(t, DB.First(&selected, 201).Error)
	assert.Equal(t, "existing-model,manual-model", selected.Models)
	var selectedAbility Ability
	require.NoError(t, DB.Where("channel_id = ? AND model = ?", 201, "manual-model").First(&selectedAbility).Error)
	assert.True(t, selectedAbility.Enabled)

	var removed Channel
	require.NoError(t, DB.First(&removed, 202).Error)
	assert.Equal(t, "other-model", removed.Models)
	var mapping map[string]string
	require.NoError(t, common.UnmarshalJsonStr(removed.GetModelMapping(), &mapping))
	assert.NotContains(t, mapping, "manual-model")
	assert.Equal(t, "other-upstream", mapping["other-model"])
	var removedAbilityCount int64
	require.NoError(t, DB.Model(&Ability{}).Where("channel_id = ? AND model = ?", 202, "manual-model").Count(&removedAbilityCount).Error)
	assert.Zero(t, removedAbilityCount)
}

func TestUpdateModelChannelBindingsAutomaticLeavesConfigurationUntouched(t *testing.T) {
	setupModelMetaTestDB(t)
	modelRow := Model{ModelName: "automatic-model"}
	require.NoError(t, DB.Create(&modelRow).Error)
	originalMapping := `{"automatic-model":"upstream-model"}`
	require.NoError(t, DB.Create(&Channel{Id: 301, Name: "channel", Key: "key-301", Models: "automatic-model", ModelMapping: &originalMapping}).Error)
	require.NoError(t, DB.Create(&Ability{Group: "default", Model: "automatic-model", ChannelId: 301, Enabled: true}).Error)

	require.NoError(t, UpdateModelChannelBindings(modelRow.Id, "auto", nil))

	var channel Channel
	require.NoError(t, DB.First(&channel, 301).Error)
	assert.Equal(t, "automatic-model", channel.Models)
	assert.Equal(t, originalMapping, channel.GetModelMapping())
	var abilityCount int64
	require.NoError(t, DB.Model(&Ability{}).Where("channel_id = ? AND model = ?", 301, "automatic-model").Count(&abilityCount).Error)
	assert.Equal(t, int64(1), abilityCount)
}

func strptr(value string) *string {
	return &value
}
