package common

import (
	"testing"

	"github.com/QuantumNous/new-api/constant"
	"github.com/stretchr/testify/require"
)

func TestValidateCodexContextTokens(t *testing.T) {
	require.NoError(t, ValidateCodexContextTokens(constant.ChannelTypeCodex, CodexMaxContextTokens))
	require.NoError(t, ValidateCodexContextTokens(constant.ChannelTypeCodex, CodexMaxContextTokens-1))
	require.ErrorContains(t, ValidateCodexContextTokens(constant.ChannelTypeCodex, CodexMaxContextTokens+1), "exceeding")
	require.NoError(t, ValidateCodexContextTokens(constant.ChannelTypeOpenAI, CodexMaxContextTokens+1))
}
