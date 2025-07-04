package utils

import (
	"encoding/json"
	"log"
)

// ToJsonString returns a pretty-formatted JSON string representation of the given value.
// It uses json.MarshalIndent to produce human-readable output.
// If marshalling fails, it returns an error describing the issue.
func ToJsonString(v interface{}) (string, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Printf("json marshal err: %v", err)
	}
	return string(data), nil
}
