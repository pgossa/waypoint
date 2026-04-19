package model

import (
	"slices"

	"waypoint/util"

	"github.com/google/uuid"
)

type Epic struct {
	id      uuid.UUID
	name    string
	path    []string
	done    bool
	tasksID []uuid.UUID
}

type EpicJSON struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Path    []string `json:"path"`
	Done    bool     `json:"done"`
	TasksID []string `json:"tasks_id"`
}

func CreateEpic(name string, path []string, tasksID []uuid.UUID) (*Epic, error) {
	if err := util.ValidateName(name); err != nil {
		return nil, err
	}

	if err := util.ValidatePath(path); err != nil {
		return nil, err
	}

	return &Epic{id: uuid.New(), name: name, path: path, done: false, tasksID: tasksID}, nil
}

func (e *Epic) GetID() uuid.UUID {
	return e.id
}

func (e *Epic) GetName() string {
	return e.name
}

func (e *Epic) GetPath() []string {
	return e.path
}

func (e *Epic) MarkDone() {
	e.done = true
}

func (e *Epic) IsDone() bool {
	return e.done
}

func (e *Epic) GetTasksID() []uuid.UUID {
	return e.tasksID
}

func (e *Epic) RemoveTaskID(taskID uuid.UUID) {
	e.tasksID = slices.DeleteFunc(e.tasksID, func(id uuid.UUID) bool {
		return id == taskID
	})
}

func (e *Epic) AddTaskID(taskID uuid.UUID) error {
	if slices.Contains(e.tasksID, taskID) {
		return util.ErrDuplicateTask
	}
	e.tasksID = append(e.tasksID, taskID)
	return nil
}

func (e *Epic) ToJSON() EpicJSON {
	tasksID := make([]string, len(e.tasksID))
	for i, id := range e.tasksID {
		tasksID[i] = id.String()
	}
	return EpicJSON{
		ID:      e.id.String(),
		Name:    e.name,
		Path:    e.path,
		Done:    e.done,
		TasksID: tasksID,
	}
}

func FromEpicJSON(e EpicJSON) (*Epic, error) {
	id, err := uuid.Parse(e.ID)
	if err != nil {
		return nil, err
	}
	tids := make([]uuid.UUID, 0, len(e.TasksID))
	for _, tid := range e.TasksID {
		parsed, err := uuid.Parse(tid)
		if err != nil {
			return nil, err
		}
		tids = append(tids, parsed)
	}
	return &Epic{id: id, name: e.Name, path: e.Path, done: e.Done, tasksID: tids}, nil
}
