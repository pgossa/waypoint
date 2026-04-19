package job

import (
	"strings"
	"testing"

	"waypoint/model"
	"waypoint/util"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Name validation (as enforced by model.CreateJob inside jobAdd)
// The allowed charset is: [a-zA-Z0-9 ._/#@-]
// =============================================================================

func TestJobAdd_ValidNames(t *testing.T) {
	valid := []string{
		"fix-something",
		"fix.something",
		"fix_something",
		"fixSomething123",
		"fix something",    // space is allowed
		"job@name",         // @ is allowed
		"job/name",         // / is allowed
		"job#name",         // # is allowed
		strings.Repeat("a", 64), // exactly 64 chars
	}
	for _, name := range valid {
		name := name
		t.Run(name, func(t *testing.T) {
			job, err := model.CreateJob(name)
			assert.NoError(t, err, "expected no error for %q", name)
			assert.NotNil(t, job)
		})
	}
}

func TestJobAdd_InvalidNames(t *testing.T) {
	tests := []struct {
		name    string
		wantErr error
	}{
		{"", util.ErrNameLength},
		{strings.Repeat("a", 65), util.ErrNameLength},
		{`job\name`, util.ErrNameInvalid},  // backslash not allowed
		{"job!name", util.ErrNameInvalid},  // ! not allowed
		{"job|name", util.ErrNameInvalid},  // | not allowed
		{"job?name", util.ErrNameInvalid},  // ? not allowed
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			_, err := model.CreateJob(tt.name)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestJobAdd_Exactly64Chars_IsValid(t *testing.T) {
	name := strings.Repeat("a", 64)
	job, err := model.CreateJob(name)
	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, name, job.GetName())
}

func TestJobAdd_65Chars_IsInvalid(t *testing.T) {
	_, err := model.CreateJob(strings.Repeat("a", 65))
	assert.Equal(t, util.ErrNameLength, err)
}

// =============================================================================
// Usage smoke test
// =============================================================================

func TestJobAddUsage_NoPanic(t *testing.T) {
	assert.NotPanics(t, func() { jobAddUsage() })
}
