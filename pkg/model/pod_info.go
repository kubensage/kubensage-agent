package model

import (
	"encoding/json"
	"fmt"
)

type PodInfo struct {
	Id          string
	Name        string
	Namespace   string
	Uid         string
	State       string
	CreatedAt   int64
	Annotations map[string]string
	Labels      map[string]string
}

func ToJsonString(p PodInfo) (string, error) {
	data, err := json.MarshalIndent(p, "", "  ")

	if err != nil {
		return "", fmt.Errorf("error marshalling PodInfo: %s", err.Error())
	}

	return string(data), nil
}
