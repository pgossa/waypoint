package storage

import (
	"encoding/json"
	"os"
	"path/filepath"

	"waypoint/config"
	"waypoint/model"
	"waypoint/util"

	"github.com/google/uuid"
)

type store struct {
	Jobs  []model.JobJSON  `json:"jobs"`
	Tasks []model.TaskJSON `json:"tasks"`
	Epics []model.EpicJSON `json:"epics"`
}

// STORE

func loadStore() (store, error) {
	path, err := config.DataPath()
	if err != nil {
		return store{}, err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return store{}, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return store{}, err
	}
	defer f.Close()

	var s store
	if err := json.NewDecoder(f).Decode(&s); err != nil {
		return store{}, err
	}
	return s, nil
}

func saveStore(s store) error {
	path, err := config.DataPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(s)
}

// JOB

func upsertJobJSON(j *model.Job) error {
	s, err := loadStore()
	if err != nil {
		return err
	}

	jobJSON := j.ToJSON()
	for i, existing := range s.Jobs {
		if existing.ID == jobJSON.ID {
			s.Jobs[i] = jobJSON
			return saveStore(s)
		}
	}

	s.Jobs = append(s.Jobs, jobJSON)
	return saveStore(s)
}

func getJobsJSON() ([]model.JobJSON, error) {
	s, err := loadStore()
	if err != nil {
		return nil, err
	}
	return s.Jobs, nil
}

func removeJobJSON(id uuid.UUID) error {
	s, err := loadStore()
	if err != nil {
		return err
	}

	for i, job := range s.Jobs {
		if job.ID == id.String() {
			s.Jobs = append(s.Jobs[:i], s.Jobs[i+1:]...)
			return saveStore(s)
		}
	}

	return util.ErrNotFound
}

// TASK

func upsertTaskJSON(t *model.Task) error {
	s, err := loadStore()
	if err != nil {
		return err
	}

	taskJSON := t.ToJSON()
	for i, existing := range s.Tasks {
		if existing.ID == taskJSON.ID {
			s.Tasks[i] = taskJSON
			return saveStore(s)
		}
	}

	s.Tasks = append(s.Tasks, taskJSON)
	return saveStore(s)
}

func getTasksJSON() ([]model.TaskJSON, error) {
	s, err := loadStore()
	if err != nil {
		return nil, err
	}
	return s.Tasks, nil
}

func removeTaskJSON(id uuid.UUID) error {
	s, err := loadStore()
	if err != nil {
		return err
	}
	for i, task := range s.Tasks {
		if task.ID == id.String() {
			s.Tasks = append(s.Tasks[:i], s.Tasks[i+1:]...)
			return saveStore(s)
		}
	}
	return util.ErrNotFound
}

// EPIC

func upsertEpicJSON(e *model.Epic) error {
	s, err := loadStore()
	if err != nil {
		return err
	}

	epicJSON := e.ToJSON()
	for i, existing := range s.Epics {
		if existing.ID == epicJSON.ID {
			s.Epics[i] = epicJSON
			return saveStore(s)
		}
	}

	s.Epics = append(s.Epics, epicJSON)
	return saveStore(s)
}

func getEpicsJSON() ([]model.EpicJSON, error) {
	s, err := loadStore()
	if err != nil {
		return nil, err
	}
	return s.Epics, nil
}

func removeEpicJSON(id uuid.UUID) error {
	s, err := loadStore()
	if err != nil {
		return err
	}
	for i, epic := range s.Epics {
		if epic.ID == id.String() {
			s.Epics = append(s.Epics[:i], s.Epics[i+1:]...)
			return saveStore(s)
		}
	}
	return util.ErrNotFound
}
