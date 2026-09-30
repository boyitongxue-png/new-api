package ratio_setting

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateRatioMaps(t *testing.T) {
	valid := `{"claude-opus-4-8":2.5,"free-model":0}`
	assert.NoError(t, ValidateModelRatioJSONString(valid))
	assert.NoError(t, ValidateModelPriceJSONString(valid))
	assert.NoError(t, ValidateCacheRatioJSONString(valid))

	for name, value := range map[string]string{
		"negative":         `{"model":-0.1}`,
		"infinite":         `{"model":1e1000}`,
		"empty model":      `{"":1}`,
		"null payload":     `null`,
		"null value":       `{"model":null}`,
		"over upper bound": `{"model":1000000001}`,
	} {
		t.Run(name, func(t *testing.T) {
			require.Error(t, ValidateModelPriceJSONString(value))
		})
	}
}

func TestValidateNestedGroupRatios(t *testing.T) {
	assert.NoError(t, ValidateGroupRatioJSONString(`{"default":1,"vip":0.8}`))
	assert.NoError(t, ValidateGroupGroupRatioJSONString(`{"vip":{"standard":0.9}}`))

	require.Error(t, ValidateGroupRatioJSONString(`{"vip":-1}`))
	require.Error(t, ValidateGroupGroupRatioJSONString(`{"vip":{"":1}}`))
	require.Error(t, ValidateGroupGroupRatioJSONString(`{"vip":{"standard":1000000001}}`))
}

func TestUpdateModelPriceRejectsInvalidPayloadWithoutReplacingMap(t *testing.T) {
	saved := ModelPrice2JSONString()
	t.Cleanup(func() { require.NoError(t, UpdateModelPriceByJSONString(saved)) })

	require.NoError(t, UpdateModelPriceByJSONString(`{"validation-model":1.25}`))
	_, ok := GetModelPrice("validation-model", false)
	require.True(t, ok)
	require.Error(t, UpdateModelPriceByJSONString(`{"validation-model":-1}`))
	price, ok := GetModelPrice("validation-model", false)
	require.True(t, ok)
	assert.Equal(t, 1.25, price)
}
