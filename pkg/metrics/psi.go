package metrics

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
)

type PsiData struct {
	Total  uint64  `json:"total,omitempty"`
	Avg10  float64 `json:"avg10,omitempty"`
	Avg60  float64 `json:"avg60,omitempty"`
	Avg300 float64 `json:"avg300,omitempty"`
}

type PsiMetrics struct {
	Resource string
	Some     PsiData
	Full     PsiData
}

func SafePsiMetrics(path string) PsiMetrics {
	file, err := os.Open(path)
	if err != nil {
		log.Printf("Failed to open metrics file: %v", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Printf("error closing %s: %v", path, err.Error())
		}
	}(file)

	scanner := bufio.NewScanner(file)
	metrics := PsiMetrics{Resource: strings.TrimPrefix(path, "/proc/pressure/")}

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) < 5 {
			continue
		}

		entry := PsiData{}
		for _, part := range parts[1:] {
			kv := strings.Split(part, "=")
			if len(kv) != 2 {
				continue
			}
			switch kv[0] {
			case "avg10":
				entry.Avg10, _ = strconv.ParseFloat(kv[1], 64)
			case "avg60":
				entry.Avg60, _ = strconv.ParseFloat(kv[1], 64)
			case "avg300":
				entry.Avg300, _ = strconv.ParseFloat(kv[1], 64)
			case "total":
				entry.Total, _ = strconv.ParseUint(kv[1], 10, 64)
			}
		}

		switch parts[0] {
		case "some":
			metrics.Some = entry
		case "full":
			metrics.Full = entry
		}
	}

	return metrics
}
