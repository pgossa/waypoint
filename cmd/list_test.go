package cmd

import (
	"testing"
)

// =============================================================================
// pathMatches — duplicate of util.PathMatches defined in list.go
// =============================================================================

func TestPathMatches_Identical(t *testing.T) {
	if !pathMatches([]string{"home", "user"}, []string{"home", "user"}) {
		t.Error("expected match for identical paths")
	}
}

func TestPathMatches_Empty(t *testing.T) {
	if !pathMatches([]string{}, []string{}) {
		t.Error("expected match for two empty paths")
	}
}

func TestPathMatches_LengthMismatch(t *testing.T) {
	tests := []struct {
		a, b []string
	}{
		{[]string{"home"}, []string{"home", "user"}},
		{[]string{"home", "user"}, []string{"home"}},
		{[]string{}, []string{"home"}},
	}
	for _, tt := range tests {
		if pathMatches(tt.a, tt.b) {
			t.Errorf("expected no match for %v vs %v", tt.a, tt.b)
		}
	}
}

func TestPathMatches_DifferentContent(t *testing.T) {
	if pathMatches([]string{"home", "alice"}, []string{"home", "bob"}) {
		t.Error("expected no match when content differs")
	}
}

func TestPathMatches_CaseSensitive(t *testing.T) {
	if pathMatches([]string{"Home"}, []string{"home"}) {
		t.Error("expected case-sensitive comparison")
	}
}
