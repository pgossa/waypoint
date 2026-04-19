package storage

import (
	"waypoint/config"
	"waypoint/model"
	"waypoint/util"

	"github.com/google/uuid"
)

// JOB

func SaveJob(j *model.Job) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	switch cfg.Storage {
	case config.StorageJSON:
		return upsertJobJSON(j)
	case config.StorageSQLite:
		return upsertJobSQLite(j)
	default:
		return util.ErrStorageNotSupported
	}
}

func GetJobs() ([]*model.Job, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	switch cfg.Storage {
	case config.StorageJSON:
		jobsJSON, err := getJobsJSON()
		if err != nil {
			return nil, err
		}
		jobs := []*model.Job{}
		for _, jobJSON := range jobsJSON {
			job, err := jobJSON.FromJobJSON()
			if err != nil {
				return nil, err
			}
			jobs = append(jobs, job)
		}
		return jobs, nil
	case config.StorageSQLite:
		jobs, err := getJobsSQLite()
		if err != nil {
			return nil, err
		}
		return jobs, nil
	default:
		return nil, util.ErrStorageNotSupported
	}
}

func RemoveJob(id uuid.UUID) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	switch cfg.Storage {
	case config.StorageJSON:
		return removeJobJSON(id)
	case config.StorageSQLite:
		return removeJobSQLite(id)
	default:
		return util.ErrStorageNotSupported
	}
}

// TASK

func SaveTask(t *model.Task) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	switch cfg.Storage {
	case config.StorageJSON:
		return upsertTaskJSON(t)
	case config.StorageSQLite:
		return upsertTaskSQLite(t)
	default:
		return util.ErrStorageNotSupported
	}
}

func GetTasks() ([]*model.Task, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	var jsons []model.TaskJSON
	var tasks []*model.Task
	switch cfg.Storage {
	case config.StorageJSON:
		jsons, err = getTasksJSON()
	case config.StorageSQLite:
		tasks, err = getTasksSQLite()
		if err != nil {
			return nil, err
		}
		return tasks, nil
	default:
		return nil, util.ErrStorageNotSupported
	}
	if err != nil {
		return nil, err
	}

	tasks = make([]*model.Task, 0, len(jsons))
	for _, j := range jsons {
		task, err := model.FromTaskJSON(j)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func RemoveTask(id uuid.UUID) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	switch cfg.Storage {
	case config.StorageJSON:
		return removeTaskJSON(id)
	case config.StorageSQLite:
		return removeTaskSQLite(id)
	default:
		return util.ErrStorageNotSupported
	}
}

// EPIC

func SaveEpic(e *model.Epic) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	switch cfg.Storage {
	case config.StorageJSON:
		return upsertEpicJSON(e)
	case config.StorageSQLite:
		return upsertEpicSQLite(e)
	default:
		return util.ErrStorageNotSupported
	}
}

func GetEpics() ([]*model.Epic, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	var epics []*model.Epic
	var jsons []model.EpicJSON
	var err2 error
	switch cfg.Storage {
	case config.StorageJSON:
		jsons, err2 = getEpicsJSON()
	case config.StorageSQLite:
		epics, err2 = getEpicsSQLite()
		if err2 != nil {
			return nil, err2
		}
		return epics, nil
	default:
		return nil, util.ErrStorageNotSupported
	}
	if err2 != nil {
		return nil, err2
	}

	epics = make([]*model.Epic, 0, len(jsons))
	for _, j := range jsons {
		epic, err := model.FromEpicJSON(j)
		if err != nil {
			return nil, err
		}
		epics = append(epics, epic)
	}
	return epics, nil
}

func RemoveEpic(id uuid.UUID) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	switch cfg.Storage {
	case config.StorageJSON:
		return removeEpicJSON(id)
	case config.StorageSQLite:
		return removeEpicSQLite(id)
	default:
		return util.ErrStorageNotSupported
	}
}
