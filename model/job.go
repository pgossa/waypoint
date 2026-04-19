package model

import (
	"waypoint/util"

	"github.com/google/uuid"
)

type Job struct {
	id   uuid.UUID
	name string
	done bool
}

type JobJSON struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Done bool   `json:"done"`
}

func CreateJob(name string) (*Job, error) {
	if err := util.ValidateName(name); err != nil {
		return nil, err
	}
	return &Job{id: uuid.New(), name: name, done: false}, nil
}

func (j *Job) GetName() string {
	return j.name
}

func (j *Job) GetID() uuid.UUID {
	return j.id
}

func (j *Job) MarkDone() {
	j.done = true
}

func (j *Job) IsDone() bool {
	return j.done
}

func (j *Job) ToJSON() JobJSON {
	return JobJSON{
		ID:   j.id.String(),
		Name: j.name,
		Done: j.done,
	}
}

func (j JobJSON) FromJobJSON() (*Job, error) {
	id, err := uuid.Parse(j.ID)
	if err != nil {
		return nil, err
	}
	return &Job{id: id, name: j.Name, done: j.Done}, nil
}
