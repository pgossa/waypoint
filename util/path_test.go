package util_test

import (
	"testing"

	"waypoint/util"
)

// =============================================================================
// GetPathSplit — smoke test (always returns non-nil in a test environment)
// =============================================================================

func TestGetPathSplit_ReturnsNonEmpty(t *testing.T) {
	parts := util.GetPathSplit()
	if len(parts) == 0 {
		t.Error("GetPathSplit returned empty slice in a running test environment")
	}
	for i, p := range parts {
		if p == "" {
			t.Errorf("GetPathSplit returned empty segment at index %d", i)
		}
	}
}

// =============================================================================
// PathMatches — exact equality of two path slices
// =============================================================================

func TestPathMatches(t *testing.T) {
	tests := []struct {
		name       string
		taskPath   []string
		activePath []string
		want       bool
	}{
		{"identical single segment", []string{"home"}, []string{"home"}, true},
		{"identical multi segment", []string{"home", "user", "projects"}, []string{"home", "user", "projects"}, true},

		// length mismatches — not a match even if one is a prefix
		{"task shorter than active", []string{"home"}, []string{"home", "user"}, false},
		{"task longer than active", []string{"home", "user"}, []string{"home"}, false},
		{"both empty", []string{}, []string{}, true},
		{"one empty", []string{}, []string{"home"}, false},

		// same length, different content
		{"same length different content", []string{"home", "alice"}, []string{"home", "bob"}, false},
		{"differs at first segment", []string{"work", "projects"}, []string{"home", "projects"}, false},
		{"differs at last segment", []string{"home", "user", "a"}, []string{"home", "user", "b"}, false},

		// case-sensitive
		{"case mismatch", []string{"Home"}, []string{"home"}, false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := util.PathMatches(tt.taskPath, tt.activePath)
			if got != tt.want {
				t.Errorf("PathMatches(%v, %v) = %v, want %v", tt.taskPath, tt.activePath, got, tt.want)
			}
		})
	}
}
