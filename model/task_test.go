package model

import (
	"testing"
	"waypoint/util"
)

func TestCreateTask(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{"valid name", "fix auth bug", nil},
		{"empty name", "", util.ErrName},
		{"name too long", string(make([]byte, 65)), util.ErrName},
		{"name at max length", string(make([]byte, 64)), nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, err := CreateTask(tt.input)
			if err != tt.wantErr {
				t.Errorf("CreateTask(%q) error = %v, want %v", tt.input, err, tt.wantErr)
			}
			if err != nil && task != nil {
				t.Errorf("expected nil task on error, got %v", task)
			}
			if err == nil && task == nil {
				t.Error("expected task, got nil")
			}
		})
	}
}
