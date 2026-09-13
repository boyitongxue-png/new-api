package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeImageResolution(t *testing.T) {
	tests := map[string]string{
		"1K":        "1k",
		"2k":        "2k",
		"4096x4096": "4k",
		"1024x1024": "1k",
		"unknown":   "",
	}
	for input, want := range tests {
		require.Equal(t, want, NormalizeImageResolution(input))
	}
}

func TestSelectLowestCostImageChannelsFiltersByResolutionAndSorts(t *testing.T) {
	channels := []*Channel{
		{Id: 3, OtherInfo: `{"supportedResolutions":["2k"],"cost2k":0.05}`},
		{Id: 1, OtherInfo: `{"supportedResolutions":["2k"],"cost2k":0.02}`},
		{Id: 2, OtherInfo: `{"supportedResolutions":["1k"],"cost1k":0.01}`},
	}
	selected, ok := SelectLowestCostImageChannels(channels, "2k")
	require.True(t, ok)
	require.Equal(t, []int{1, 3}, []int{selected[0].Id, selected[1].Id})
}

func TestSelectLowestCostImageChannelsFallsBackWithoutCost(t *testing.T) {
	channels := []*Channel{{Id: 1, OtherInfo: `{"supportedResolutions":["2k"]}`}}
	selected, ok := SelectLowestCostImageChannels(channels, "2k")
	require.False(t, ok)
	require.Nil(t, selected)
}