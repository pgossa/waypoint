package job

import (
	"fmt"
	"os"

	"waypoint/ui"
)

func Job(args []string) {
	if len(args) < 1 {
		jobUsage()
		return
	}
	command, rest := args[0], args[1:]
	switch command {
	case "add", "a":
		jobAdd(rest)
	case "list", "ls", "l":
		jobList(rest)
	case "done", "d":
		jobDone(rest)
	case "remove", "r":
		jobRemove(rest)
	case "help", "--help", "-h":
		jobUsage()
	default:
		ui.Info(fmt.Sprintf("wpt job: unknown command %q\n\n", command))
		jobUsage()
		os.Exit(1)
	}
}

func jobUsage() {
}
