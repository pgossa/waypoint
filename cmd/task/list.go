package task

import (
	"os"
	"strings"

	"waypoint/model"
	"waypoint/ui"
	"waypoint/util"
)

func taskList(args []string) {
	var customPath []string
	showAll := false
	showDone := false

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--all", "-a", "all":
			showAll = true
		case "--done", "-d":
			showDone = true
		case "--path", "-p":
			if i+1 >= len(args) {
				ui.Error("--path requires an argument")
				os.Exit(1)
			}
			i++
			customPath = strings.Split(args[i], "/")
		default:
			ui.Error("unknown flag: " + args[i])
			taskListUsage()
			os.Exit(1)
		}
	}

	var tasks []*model.Task
	var err error
	if showDone {
		tasks, err = getSortedTasks()
	} else {
		tasks, err = getActiveTasks()
	}
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	activePath := customPath
	if activePath == nil && !showAll {
		activePath = util.GetPathSplit()
	}

	filtered := make([]*model.Task, 0)
	for _, task := range tasks {
		if !showAll && !util.PathMatches(task.GetPath(), activePath) {
			continue
		}
		filtered = append(filtered, task)
	}

	if len(filtered) == 0 {
		ui.Muted("No tasks found.")
		return
	}

	items := make([]ui.ListItem, len(filtered))
	for i, task := range filtered {
		index := i + 1
		if showAll || showDone {
			index = 0 // no index in flat views
		}

		sub := ""
		if task.GetEpicID() != nil {
			sub = "part of an epic"
		}

		items[i] = ui.ListItem{
			Index: index,
			Done:  task.IsDone(),
			Name:  task.GetName(),
			Path:  strings.Join(task.GetPath(), "/"),
			Sub:   sub,
		}
	}
	ui.List("TASK", items)
}

func taskListUsage() {
	ui.Info("Usage: wpt task list [flags]")
	ui.Info("  --all,  -a        Show tasks from all paths")
	ui.Info("  --done, -d        Include completed tasks")
	ui.Info("  --path, -p <path> Filter by specific path")
}
