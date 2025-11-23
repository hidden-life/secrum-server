package real_time

import "encoding/json"

type OutEnvelope struct {
	Type string `json:"type"`
	Data any    `json:"data,omitempty"`
}

type InEnvelope struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}
