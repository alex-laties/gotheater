package message

import "encoding/json"

// Base represents the basic structure of a gotheater message used for parsing
type Base struct {
	ID   string          `json:"id"`
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// SetRuler represents the data payload of a setRuler message
type SetRuler struct {
	NewRulerID string `json:"newRulerID"`
}

// SetMedia represents the data payload of a setMedia message
type SetMedia struct {
	URL string `json:"url"`
}

// Seek represents the data payload of a seek message
type Seek struct {
	MediaTimestamp int `json:"mediaTimestamp"`
}

// Status represents the data payload of a status message
type Status struct {
	Name                  string  `json:"name"`
	Playing               bool    `json:"playing"`
	CurrentMediaURL       string  `json:"currentMediaURL"`
	CurrentMediaTimestamp int     `json:"currentMediaTimestamp"`
	CurrentPing           int     `json:"currentPing"`
	CurrentPlaybackRate   float64 `json:"currentPlaybackRate"`
}

// RulerPlaybackStatus represents the data payload of a playback status message
type RulerPlaybackStatus struct {
	Playing               bool `json:"playing"`
	CurrentMediaTimestamp int  `json:"currentMediaTimestamp"`
	CurrentPing           int  `json:"currentPing"`
}

// Ping represents the data payload of a ping message
type Ping struct {
	Timestamp int `json:"timestamp"`
}

// Pong represents the data payload of a pong message
type Pong struct {
	ReceivedAt int `json:"receivedAt"`
	Ping
}

// NewConnect constructs a full Connect message
func NewConnect(id string, data map[string]interface{}) map[string]interface{} {
	if data == nil {
		data = make(map[string]interface{})
	}

	if _, exists := data["id"]; !exists {
		data["id"] = id
	}

	return map[string]interface{}{
		"id":   "god",
		"type": "connect",
		"data": data,
	}
}

// NewRuler constructs a full NewRuler message
func NewRuler(id string) map[string]interface{} {
	return map[string]interface{}{
		"id":   "god",
		"type": "setRuler",
		"data": SetRuler{
			NewRulerID: id,
		},
	}
}

// NewDisconnect constructs a full Disconnect message
func NewDisconnect(id string) map[string]interface{} {
	return map[string]interface{}{
		"id":   "god",
		"type": "disconnect",
		"data": map[string]string{
			"id": id,
		},
	}
}
