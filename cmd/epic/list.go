package epic

import (
	"os"
	"strings"

	"waypoint/model"
	"waypoint/storage"
	"waypoint/ui"
	"waypoint/util"

	"github.com/google/uuid"
)

func epicList(args []string) {
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
			epicListUsage()
			os.Exit(1)
		}
	}

	var epics []*model.Epic
	var err error
	if showDone {
		epics, err = GetSortedEpics()
	} else {
		epics, err = GetActiveEpics()
	}
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	activePath := customPath
	if activePath == nil && !showAll {
		activePath = util.GetPathSplit()
	}

	filtered := make([]*model.Epic, 0)
	for _, epic := range epics {
		if !showAll && !util.PathMatches(epic.GetPath(), activePath) {
			continue
		}
		filtered = append(filtered, epic)
	}

	if len(filtered) == 0 {
		ui.Muted("No epics found.")
		return
	}

	// fetch all tasks once
	allTasks, err := storage.GetTasks()
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	taskMap := make(map[uuid.UUID]*model.Task)
	for _, t := range allTasks {
		taskMap[t.GetID()] = t
	}

	items := make([]ui.ListItem, 0, len(filtered))
	for i, epic := range filtered {
		taskIDs := epic.GetTasksID()
		total := len(taskIDs)
		doneCount := 0
		for _, tid := range taskIDs {
			if t, ok := taskMap[tid]; ok && t.IsDone() {
				doneCount++
			}
		}

		subLines := make([]string, 0)
		if total > 0 {
			subLines = append(subLines, ui.ProgressBar(doneCount, total))
			for _, tid := range taskIDs {
				t, ok := taskMap[tid]
				if !ok {
					continue
				}
				var prefix string
				if t.IsDone() {
					prefix = "✓ "
				} else {
					prefix = "◆ "
				}
				subLines = append(subLines, prefix+t.GetName()+"  "+strings.Join(t.GetPath(), "/"))
			}
		} else {
			subLines = append(subLines, "no subtasks")
		}

		item := ui.ListItem{
			Done: epic.IsDone(),
			Name: epic.GetName(),
			Sub:  strings.Join(subLines, "\n"),
		}
		if showAll || showDone {
			item.Path = strings.Join(epic.GetPath(), "/")
		} else {
			item.Index = i + 1
		}
		items = append(items, item)
	}

	ui.List("EPIC", items)
}

func epicListUsage() {
	ui.Info("Usage: wpt epic list [flags]")
	ui.Info("  --all,  -a        Show epics from all paths")
	ui.Info("  --done, -d        Include completed epics")
	ui.Info("  --path, -p <path> Filter by specific path")
}
