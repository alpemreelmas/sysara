package utils

import (
	"fmt"
)

// FormatBytes converts bytes to human readable format
func FormatBytes(bytes uint64) string {
	if bytes == 0 {
		return "0 Bytes"
	}

	const unit = 1024
	sizes := []string{"Bytes", "KB", "MB", "GB", "TB", "PB"}

	b := float64(bytes)
	var i int
	for b >= unit && i < len(sizes)-1 {
		b /= unit
		i++
	}

	return fmt.Sprintf("%.1f %s", b, sizes[i])
}

// FormatUptime converts seconds to human readable uptime format
func FormatUptime(seconds uint64) string {
	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	minutes := (seconds % 3600) / 60

	return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
}

// GetStatusClass returns CSS class based on process status
func GetStatusClass(status string) string {
	switch status {
	case "running":
		return "bg-green-100 text-green-800"
	case "sleeping":
		return "bg-blue-100 text-blue-800"
	case "stopped":
		return "bg-red-100 text-red-800"
	case "zombie":
		return "bg-yellow-100 text-yellow-800"
	default:
		return "bg-gray-100 text-gray-800"
	}
}
