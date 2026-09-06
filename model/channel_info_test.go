package model

import (
	"testing"

	"github.com/QuantumNous/new-api/constant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChannelInfoValueAndScanSupportPostgresJSONRepresentations(t *testing.T) {
	expected := ChannelInfo{
		IsMultiKey:         true,
		MultiKeySize:       2,
		MultiKeyStatusList: map[int]int{0: 1, 1: 0},
		MultiKeyMode:       constant.MultiKeyModePolling,
	}

	value, err := expected.Value()
	require.NoError(t, err)
	jsonText, ok := value.(string)
	require.True(t, ok, "ChannelInfo.Value must return JSON text, got %T", value)

	for _, test := range []struct {
		name  string
		value interface{}
	}{
		{name: "database string", value: jsonText},
		{name: "database bytes", value: []byte(jsonText)},
	} {
		t.Run(test.name, func(t *testing.T) {
			var got ChannelInfo
			require.NoError(t, got.Scan(test.value))
			assert.Equal(t, expected, got)
		})
	}
}

func TestChannelInfoScanTreatsNullValuesAsEmpty(t *testing.T) {
	for _, value := range []interface{}{nil, []byte(" "), " ", []byte("null"), "null"} {
		var got ChannelInfo
		require.NoError(t, got.Scan(value))
		assert.Equal(t, ChannelInfo{}, got)
	}
}

func TestChannelInfoScanRejectsInvalidValueTypeAndJSON(t *testing.T) {
	var info ChannelInfo

	err := info.Scan(123)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported channel_info scan value type")

	err = info.Scan([]byte("{"))
	require.Error(t, err)
}
