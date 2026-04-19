package task

import (
	"fmt"
	"os"

	"waypoint/ui"
)

func Task(args []string) {
	if len(args) < 1 {
		taskUsage()
		return
	}
	command, rest := args[0], args[1:]
	switch command {
	case "add", "a":
		taskAdd(rest)
	case "bound", "b":
		taskBound(rest)
	case "done", "d":
		taskDone(rest)
	case "list", "ls", "l":
		taskList(rest)
	case "remove", "r":
		taskRemove(rest)
	case "unbound", "u":
		taskUnbound(rest)
	case "help", "--help", "-h":
		taskUsage()
	default:
		ui.Info(fmt.Sprintf("wpt task: unknown command %q\n\n", command))
		taskUsage()
		os.Exit(1)
	}
}

func taskUsage() {
}
