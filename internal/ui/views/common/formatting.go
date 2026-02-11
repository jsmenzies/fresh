package common

import (
	"fmt"
	"time"
)

func FormatTimeAgo(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}

	duration := time.Since(t)

	if duration < time.Minute {
		return "just now"
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "< 2m"
		}
		return fmt.Sprintf("< %dm", minutes)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", hours)
	} else if duration < 7*24*time.Hour {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day "
		}
		return fmt.Sprintf("%d days", days)
	} else if duration < 30*24*time.Hour {
		weeks := int(duration.Hours() / (24 * 7))
		if weeks == 1 {
			return "1 week"
		}
		return fmt.Sprintf("%d weeks", weeks)
	} else if duration < 365*24*time.Hour {
		months := int(duration.Hours() / (24 * 30))
		if months == 1 {
			return "1 month"
		}
		return fmt.Sprintf("%d months", months)
	}

	years := int(duration.Hours() / (24 * 365))
	if years == 1 {
		return "1 year"
	}
	return fmt.Sprintf("%d years", years)
}
