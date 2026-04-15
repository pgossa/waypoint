package model

import (
	"waypoint/util"

	"github.com/google/uuid"
)

type Task struct {
	id   uuid.UUID
	name string
}

func CreateTask(name string) (*Task, error) {
	newUuid := uuid.New()
	if name == "" || len(name) > 64 {
		return nil, util.ErrName
	}
	newTask := Task{
		id:   newUuid,
		name: name,
	}
	return &newTask, nil
}
