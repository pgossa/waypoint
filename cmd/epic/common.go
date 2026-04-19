package epic

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"waypoint/model"
	"waypoint/storage"
	"waypoint/ui"
)

func GetSortedEpics() ([]*model.Epic, error) {
	epics, err := storage.GetEpics()
	if err != nil {
		return nil, err
	}
	sort.Slice(epics, func(i, j int) bool {
		return epics[i].GetName() < epics[j].GetName()
	})
	return epics, nil
}

func GetActiveEpics() ([]*model.Epic, error) {
	epics, err := GetSortedEpics()
	if err != nil {
		return nil, err
	}
	active := make([]*model.Epic, 0)
	for _, epic := range epics {
		if !epic.IsDone() {
			active = append(active, epic)
		}
	}
	return active, nil
}

func FindEpic(input string, epics []*model.Epic) (int, []int) {
	if n, err := strconv.Atoi(input); err == nil {
		if n >= 1 && n <= len(epics) {
			return n - 1, []int{n - 1}
		}
		return -1, nil
	}
	matches := make([]int, 0)
	for i, epic := range epics {
		if strings.Contains(strings.ToLower(epic.GetName()), strings.ToLower(input)) {
			matches = append(matches, i)
		}
	}
	if len(matches) == 1 {
		return matches[0], matches
	}
	return -1, matches
}

func PrintEpicWithIndex(i int, epic *model.Epic) {
	status := " "
	if epic.IsDone() {
		status = "✓"
	}
	ui.Info(fmt.Sprintf("[%d] %s %-30s /%s", i+1, status, epic.GetName(), strings.Join(epic.GetPath(), "/")))
}
