package task

import (
	"fmt"
	"sort"
	"strings"

	"waypoint/model"
	"waypoint/storage"
	"waypoint/ui"
)

func pathHasPrefix(taskPath []string, prefix []string) bool {
	if len(taskPath) < len(prefix) {
		return false
	}
	for i := range prefix {
		if taskPath[i] != prefix[i] {
			return false
		}
	}
	return true
}

func getSortedTasks() ([]*model.Task, error) {
	tasks, err := storage.GetTasks()
	if err != nil {
		return nil, err
	}
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].GetName() < tasks[j].GetName()
	})
	return tasks, nil
}

func getActiveTasks() ([]*model.Task, error) {
	tasks, err := getSortedTasks()
	if err != nil {
		return nil, err
	}
	active := make([]*model.Task, 0)
	for _, task := range tasks {
		if !task.IsDone() {
			active = append(active, task)
		}
	}
	return active, nil
}

func findTask(input string, tasks []*model.Task) (int, []int) {
	matches := make([]int, 0)
	for i, task := range tasks {
		if strings.Contains(strings.ToLower(task.GetName()), strings.ToLower(input)) {
			matches = append(matches, i)
		}
	}
	if len(matches) == 1 {
		return matches[0], matches
	}
	return -1, matches
}

func printTaskWithIndex(i int, task *model.Task) {
	status := " "
	if task.IsDone() {
		status = "✓"
	}
	ui.Info(fmt.Sprintf("[%d] %s %s", i+1, status, task.GetName()))
}
