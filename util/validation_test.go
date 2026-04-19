package util_test

import (
	"strings"
	"testing"

	"waypoint/util"
)

// =============================================================================
// ValidateName
// Allowed: [a-zA-Z0-9 ._/#@-], length 1–64
// =============================================================================

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		// length boundaries
		{"empty", "", util.ErrNameLength},
		{"1 char", "a", nil},
		{"64 chars", strings.Repeat("a", 64), nil},
		{"65 chars", strings.Repeat("a", 65), util.ErrNameLength},

		// allowed special characters
		{"space", "fix something", nil},
		{"dot", "fix.something", nil},
		{"underscore", "fix_something", nil},
		{"dash", "fix-something", nil},
		{"slash", "fix/something", nil},
		{"hash", "fix#something", nil},
		{"at", "fix@something", nil},
		{"mixed", "fix #1 @user/thing.v2-final_ok", nil},

		// disallowed characters
		{"backslash", `fix\thing`, util.ErrNameInvalid},
		{"pipe", "fix|thing", util.ErrNameInvalid},
		{"bang", "fix!thing", util.ErrNameInvalid},
		{"question mark", "fix?thing", util.ErrNameInvalid},
		{"semicolon", "fix;thing", util.ErrNameInvalid},
		{"newline", "fix\nthing", util.ErrNameInvalid},
		{"tab", "fix\tthing", util.ErrNameInvalid},
		{"percent", "fix%thing", util.ErrNameInvalid},
		{"caret", "fix^thing", util.ErrNameInvalid},
		{"ampersand", "fix&thing", util.ErrNameInvalid},
		{"asterisk", "fix*thing", util.ErrNameInvalid},
		{"paren open", "fix(thing", util.ErrNameInvalid},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := util.ValidateName(tt.input)
			if err != tt.wantErr {
				t.Errorf("ValidateName(%q) = %v, want %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

// =============================================================================
// ValidatePath
// Each segment: [a-zA-Z0-9 ._-], length 1–64; slice must be non-empty.
// Note: path segments have a stricter charset than names (no /@#).
// =============================================================================

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name    string
		path    []string
		wantErr error
	}{
		// nil / empty slice
		{"nil path", nil, util.ErrPath},
		{"empty slice", []string{}, util.ErrPath},

		// segment length
		{"empty segment", []string{""}, util.ErrPath},
		{"1-char segment", []string{"a"}, nil},
		{"64-char segment", []string{strings.Repeat("a", 64)}, nil},
		{"65-char segment", []string{strings.Repeat("a", 65)}, util.ErrPath},

		// multi-segment valid
		{"two segments", []string{"home", "user"}, nil},
		{"three segments", []string{"home", "user", "projects"}, nil},

		// one segment with allowed chars (no / @ # in segments)
		{"segment with dot", []string{"my.project"}, nil},
		{"segment with dash", []string{"my-project"}, nil},
		{"segment with underscore", []string{"my_project"}, nil},
		{"segment with space", []string{"my project"}, nil},

		// disallowed chars in segment
		{"segment with slash", []string{"root/child"}, util.ErrPathInvalid},
		{"segment with at", []string{"root@host"}, util.ErrPathInvalid},
		{"segment with hash", []string{"root#1"}, util.ErrPathInvalid},
		{"segment with pipe", []string{"root|child"}, util.ErrPathInvalid},
		{"segment with bang", []string{"root!child"}, util.ErrPathInvalid},

		// empty segment in multi-segment path
		{"empty first segment", []string{"", "child"}, util.ErrPath},
		{"empty last segment", []string{"root", ""}, util.ErrPath},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			err := util.ValidatePath(tt.path)
			if err != tt.wantErr {
				t.Errorf("ValidatePath(%v) = %v, want %v", tt.path, err, tt.wantErr)
			}
		})
	}
}
