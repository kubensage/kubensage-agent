package utils

import (
	"encoding/json"
	"fmt"
)

func ToJsonString(v interface{}) (string, error) {
	data, err := json.MarshalIndent(v, "", "  ")

	if err != nil {
		return "", fmt.Errorf("error marshalling PodInfo: %s", err.Error())
	}

	return string(data), nil
}
