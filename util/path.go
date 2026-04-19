package util

import (
	"os"
	"path/filepath"
	"strings"

	"waypoint/ui"
)

func getPath() string {
	path, err := os.Getwd()
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}
	return path
}

func GetPathSplit() []string {
	path := filepath.ToSlash(getPath())
	parts := strings.Split(path, "/")

	// remove empty strings from leading slash or drive letter (C:)
	filtered := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

func PathMatches(taskPath []string, activePath []string) bool {
	if len(taskPath) != len(activePath) {
		return false
	}
	for i := range taskPath {
		if taskPath[i] != activePath[i] {
			return false
		}
	}
	return true
}
