package cli

import (
	"fmt"
	"time"
)

func formatTimestamp(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Local().Format(time.RFC3339)
}

func formatDuration(d time.Duration) string {
	return d.String()
}

func formatExpiry(d time.Duration) string {
	if d < 0 {
		return "EXPIRED"
	}

	days := int(d / (24 * time.Hour))
	hours := int((d % (24 * time.Hour)) / time.Hour)

	return fmt.Sprintf("%dd %dh", days, hours)
}
