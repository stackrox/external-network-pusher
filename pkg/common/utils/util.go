package utils

import (
	"log"
	"strings"
)

// ToCompoundName accepts list of tags and returns a compound name.
// It ignores any empty tag (i.e.: empty string)
// If the final element only contains one string, then that string
// is returned as the compound name
func ToCompoundName(delim string, tags ...string) string {
	if delim == "" {
		delim = "/"
	}
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
	return strings.Join(filtered, delim)
}

// ToTags splits the compound name to a list of individual tags (names).
func ToTags(delim, compoundName string) []string {
	if delim == "" {
		delim = "/"
	}
	return strings.Split(compoundName, delim)
}

// StrSliceRemove removes an element from a string slice at the specified index
func StrSliceRemove(in []string, i int) []string {
	if i < 0 || i >= len(in) {
		log.Panicf("Index out of bound: %d", i)
	}
	in[i] = in[len(in)-1]
	return in[:len(in)-1]
}
