package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func useOptionUpdateTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&Option{}))

	previousDB := DB
	common.OptionMapRWMutex.Lock()
	previousMap := common.OptionMap
	common.OptionMap = map[string]string{"test-option": "before"}
	common.OptionMapRWMutex.Unlock()
	DB = db
	t.Cleanup(func() {
		DB = previousDB
		common.OptionMapRWMutex.Lock()
		common.OptionMap = previousMap
		common.OptionMapRWMutex.Unlock()
	})
	return db
}

func TestUpdateOptionDoesNotPublishWhenLookupFails(t *testing.T) {
	db := useOptionUpdateTestDB(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())

	require.Error(t, UpdateOption("test-option", "after"))
	common.OptionMapRWMutex.RLock()
	defer common.OptionMapRWMutex.RUnlock()
	assert.Equal(t, "before", common.OptionMap["test-option"])
}

func TestUpdateOptionDoesNotPublishWhenSaveFails(t *testing.T) {
	db := useOptionUpdateTestDB(t)
	require.NoError(t, db.Create(&Option{Key: "test-option", Value: "before"}).Error)
	require.NoError(t, db.Exec(`
		CREATE TRIGGER reject_option_update
		BEFORE UPDATE ON options
		BEGIN
			SELECT RAISE(FAIL, 'option update rejected');
		END
	`).Error)

	require.Error(t, UpdateOption("test-option", "after"))
	var stored Option
	require.NoError(t, db.First(&stored, "key = ?", "test-option").Error)
	assert.Equal(t, "before", stored.Value)
	common.OptionMapRWMutex.RLock()
	defer common.OptionMapRWMutex.RUnlock()
	assert.Equal(t, "before", common.OptionMap["test-option"])
}
