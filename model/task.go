package model

import (
	"waypoint/util"

	"github.com/google/uuid"
)

type Task struct {
	id     uuid.UUID
	name   string
	path   []string
	done   bool
	epicID *uuid.UUID
}
type TaskJSON struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Path   []string `json:"path"`
	Done   bool     `json:"done"`
	EpicID *string  `json:"epic_id"`
}

func CreateTask(name string, path []string, epicID ...uuid.UUID) (*Task, error) {
	var eID *uuid.UUID
	if len(epicID) == 1 {
		eID = &epicID[0]
	} else if len(epicID) > 1 {
		return nil, util.ErrEpicUUID
	}

	if err := util.ValidateName(name); err != nil {
		return nil, err
	}

	if err := util.ValidatePath(path); err != nil {
		return nil, err
	}

	return &Task{id: uuid.New(), name: name, path: path, done: false, epicID: eID}, nil
}

func (t *Task) GetName() string {
	return t.name
}

func (t *Task) GetID() uuid.UUID {
	return t.id
}

func (t *Task) GetPath() []string {
	return t.path
}

func (t *Task) MarkDone() {
	t.done = true
}

func (t *Task) IsDone() bool {
	return t.done
}

func (t *Task) SetEpicID(eID uuid.UUID) {
	t.epicID = &eID
}

func (t *Task) UnsetEpicID() {
	t.epicID = nil
}

func (t *Task) GetEpicID() *uuid.UUID {
	return t.epicID
}

func (t *Task) ToJSON() TaskJSON {
	var epicID *string
	if t.epicID != nil {
		s := t.epicID.String()
		epicID = &s
	}
	return TaskJSON{
		ID:     t.id.String(),
		Name:   t.name,
		Path:   t.path,
		Done:   t.done,
		EpicID: epicID,
	}
}

func FromTaskJSON(j TaskJSON) (*Task, error) {
	id, err := uuid.Parse(j.ID)
	if err != nil {
		return nil, err
	}
	var epicID *uuid.UUID
	if j.EpicID != nil {
		id, err := uuid.Parse(*j.EpicID)
		if err != nil {
			return nil, err
		}
		epicID = &id
	}
	return &Task{id: id, name: j.Name, path: j.Path, done: j.Done, epicID: epicID}, nil
}
