package utils

import (
	"strings"
)

// ToCompoundName accepts list of tags and returns a compound name.
// It ignores any empty tag (i.e.: empty string)
// If the final element only contains one string, then that string
// is returned as the compound name
func ToCompoundName(tags []string) string {
	filtered := make([]string, 0, len(tags))
	for _, tag := range tags {
		if tag != "" {
			filtered = append(filtered, tag)
		}
	}
	if len(filtered) == 0 {
		return ""
	}
	if len(filtered) == 1 {
		return filtered[0]
	}
	return strings.Join(filtered, "-")
}
