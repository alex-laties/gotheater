package message

import "encoding/json"

// Base represents the basic structure of a gotheater message used for parsing
type Base struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// SetMedia represents the data payload of a setMedia message
type SetMedia struct {
	URL string `json:"url"`
}

// NewConnect constructs a Connect message
func NewConnect(id, rulerID, currMedia string) map[string]interface{} {
	return map[string]interface{}{
		"type": "connect",
		"data": map[string]string{
			"id":           id,
			"rulerID":      rulerID,
			"currentMedia": currMedia,
		},
	}
}

// NewDisconnect constructs a Disconnect message
func NewDisconnect(id string) map[string]interface{} {
	return map[string]interface{}{
		"type": "disconnect",
		"data": map[string]string{
			"id": id,
		},
	}
}
