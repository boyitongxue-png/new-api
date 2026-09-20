package common

import (
	"fmt"

	"github.com/QuantumNous/new-api/constant"
)

// CodexMaxContextTokens is the estimated input-context ceiling for normal
// Codex Responses requests. The compaction endpoint is intentionally handled
// separately so an oversized context can still be compacted.
const CodexMaxContextTokens = 242000

// ValidateCodexContextTokens rejects an estimated Codex context above the
// configured ceiling. A previous_response_id may refer to server-side context
// that is not visible to the gateway, so callers should treat this as a local
// request guard rather than an upstream usage guarantee.
func ValidateCodexContextTokens(channelType, estimatedTokens int) error {
	if channelType != constant.ChannelTypeCodex || estimatedTokens <= CodexMaxContextTokens {
		return nil
	}
	return fmt.Errorf("estimated Codex context is %d tokens, exceeding the %d-token limit", estimatedTokens, CodexMaxContextTokens)
}
