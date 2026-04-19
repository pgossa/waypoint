package task

import (
	"os"
	"strings"

	"waypoint/model"
	"waypoint/storage"
	"waypoint/ui"
	"waypoint/util"

	"github.com/google/uuid"
)

func taskAdd(args []string) {
	if len(args) < 1 {
		taskAddUsage()
		os.Exit(1)
	}

	name := args[0]
	var epicID *string
	var path []string

	// parse flags
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--epic", "-e":
			if i+1 >= len(args) {
				ui.Error("--epic requires an argument")
				os.Exit(1)
			}
			if epicID != nil {
				ui.Error("only one --epic flag is allowed")
				os.Exit(1)
			}
			i++
			epicID = &args[i]
		case "--path", "-p":
			if i+1 >= len(args) {
				ui.Error("--path requires an argument")
				os.Exit(1)
			}
			i++
			path = strings.Split(args[i], "/")
		default:
			ui.Error("unknown flag: " + args[i])
			taskAddUsage()
			os.Exit(1)
		}
	}

	// default to current path
	if path == nil {
		path = util.GetPathSplit()
	}

	// parse epic UUID if provided
	var epicUUID []uuid.UUID
	if epicID != nil {
		id, err := uuid.Parse(*epicID)
		if err != nil {
			ui.Error("invalid epic UUID: " + *epicID)
			os.Exit(1)
		}
		epicUUID = []uuid.UUID{id}
	}

	task, err := model.CreateTask(name, path, epicUUID...)
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	if err := storage.SaveTask(task); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	ui.Success("task added: " + task.GetName())
}

func taskAddUsage() {
}
