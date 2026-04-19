package epic

import (
	"fmt"
	"os"
	"strings"

	"waypoint/model"
	"waypoint/storage"
	"waypoint/ui"
	"waypoint/util"
)

func epicAdd(args []string) {
	if len(args) < 1 {
		epicAddUsage()
		return
	}

	name := args[0]
	var customPath []string
	empty := false

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--empty", "-e":
			empty = true
		case "--path", "-p":
			if i+1 >= len(args) {
				ui.Error("--path requires an argument")
				os.Exit(1)
			}
			i++
			customPath = strings.Split(args[i], "/")
		default:
			ui.Error("unknown flag: " + args[i])
			epicAddUsage()
			os.Exit(1)
		}
	}

	path := customPath
	if path == nil {
		path = util.GetPathSplit()
	}

	// create epic first with no tasks
	epic, err := model.CreateEpic(name, path, nil)
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	if !empty {
		// scan one level deep subdirectories
		currentPath := strings.Join(path, "/")
		entries, err := os.ReadDir("/" + currentPath)
		if err != nil {
			ui.Error("failed to read directory: " + err.Error())
			os.Exit(1)
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			subPath := append(append([]string{}, path...), entry.Name())
			taskName := "[subtask] " + name

			task, err := model.CreateTask(taskName, subPath, epic.GetID())
			if err != nil {
				ui.Error(err.Error())
				os.Exit(1)
			}

			if err := epic.AddTaskID(task.GetID()); err != nil {
				ui.Error(err.Error())
				os.Exit(1)
			}

			if err := storage.SaveTask(task); err != nil {
				ui.Error(err.Error())
				os.Exit(1)
			}

			ui.Info(fmt.Sprintf("subtask created: %s at %s", taskName, strings.Join(subPath, "/")))
		}
	}

	if err := storage.SaveEpic(epic); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	ui.Success(fmt.Sprintf("epic added: %s", name))
}

func epicAddUsage() {
	ui.Info("Usage: wpt epic add <name> [flags]")
	ui.Info("  --empty, -e        Create epic without subtasks")
	ui.Info("  --path,  -p <path> Use specific path instead of current")
}
