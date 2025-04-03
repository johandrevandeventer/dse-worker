package dse890

import (
	"encoding/json"
	"fmt"

	"github.com/johandrevandeventer/dse-worker/internal/workers/types"
)

// Decoder processes DSE890 payloads
func Decoder(payload json.RawMessage) (decodedPayloadInfo *types.DecodedPayloadInfo, err error) {
	var data map[string]map[string]map[string]any
	if err := json.Unmarshal(payload, &data); err != nil {
		return decodedPayloadInfo, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	if len(data) == 0 {
		return decodedPayloadInfo, fmt.Errorf("empty payload")
	}

	return &types.DecodedPayloadInfo{
		RawPayload: payload,
	}, nil
}
