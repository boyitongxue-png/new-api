/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.
*/
package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestUpdateOptionsBulkPersistsEmptyModelPricingMaps(t *testing.T) {
	previousDB := DB
	previousType := common.MainDatabaseType()
	previousOptionMap := common.OptionMap
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&Option{}))
	DB = db
	common.SetMainDatabaseType(common.DatabaseTypeSQLite)
	common.OptionMap = map[string]string{}
	ratio_setting.InitRatioSettings()
	t.Cleanup(func() {
		DB = previousDB
		common.SetMainDatabaseType(previousType)
		common.OptionMap = previousOptionMap
	})

	require.NoError(t, UpdateOptionsBulk(map[string]string{
		"ModelPrice": `{"removed-model":0.1}`,
		"ModelRatio": `{"removed-model":2}`,
	}))
	require.NoError(t, UpdateOptionsBulk(map[string]string{
		"ModelPrice": `{}`,
		"ModelRatio": `{}`,
	}))

	var options []Option
	require.NoError(t, db.Find(&options).Error)
	require.Len(t, options, 2)
	for _, option := range options {
		require.Equal(t, "{}", option.Value)
	}
	_, found := ratio_setting.GetModelPrice("removed-model", false)
	require.False(t, found)
	_, found = ratio_setting.GetModelRatioCopy()["removed-model"]
	require.False(t, found)
}
