package storage

import (
	"waypoint/model"
	"waypoint/util"

	"github.com/google/uuid"
)

// JOB

func upsertJobSQLite(j *model.Job) error {
	return util.ErrNotImplemented
}

func getJobsSQLite() ([]*model.Job, error) {
	return nil, util.ErrNotImplemented
}

func removeJobSQLite(id uuid.UUID) error {
	return util.ErrNotImplemented
}

// TASK

func upsertTaskSQLite(t *model.Task) error {
	return util.ErrNotImplemented
}

func getTasksSQLite() ([]*model.Task, error) {
	return nil, util.ErrNotImplemented
}

func removeTaskSQLite(id uuid.UUID) error {
	return util.ErrNotImplemented
}

// EPIC

func upsertEpicSQLite(e *model.Epic) error {
	return util.ErrNotImplemented
}

func getEpicsSQLite() ([]*model.Epic, error) {
	return nil, util.ErrNotImplemented
}

func removeEpicSQLite(id uuid.UUID) error {
	return util.ErrNotImplemented
}
