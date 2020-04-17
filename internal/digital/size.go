package digital

import (
	"fmt"
	"strings"
)

const (
	sizeByte = 1.0 << (10 * iota)
	sizeKb
	sizeMb
	sizeGb
	sizeTb
)

func FormatBytes(bytes uint64) string {
	unit := ""
	value := float32(bytes)

	switch {
	case bytes >= sizeTb:
		unit = "TB"
		value = value / sizeTb
	case bytes >= sizeGb:
		unit = "GB"
		value = value / sizeGb
	case bytes >= sizeMb:
		unit = "MB"
		value = value / sizeMb
	case bytes >= sizeKb:
		unit = "KB"
		value = value / sizeKb
	case bytes >= sizeByte:
		unit = "B"
	case bytes == 0:
		return "0"
	}

	stringValue := fmt.Sprintf("%.1f", value)
	stringValue = strings.TrimSuffix(stringValue, ".0")
	return fmt.Sprintf("%s%s", stringValue, unit)
}
