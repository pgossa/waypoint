package epic

import (
	"fmt"
	"os"

	"waypoint/ui"
)

func Epic(args []string) {
	if len(args) < 1 {
		epicUsage()
		return
	}

	command, rest := args[0], args[1:]
	switch command {
	case "add", "a":
		epicAdd(rest)
	case "list", "ls", "l":
		epicList(rest)
	case "done", "d":
		epicDone(rest)
	case "remove", "rm", "r":
		epicRemove(rest)
	case "unbound", "u":
		epicUnbound(rest)
	case "help", "--help", "-h":
		epicUsage()
	default:
		ui.Error(fmt.Sprintf("wpt epic: unknown command %q", command))
		epicUsage()
		os.Exit(1)
	}
}

func epicUsage() {
	ui.Info("Usage: wpt epic <command>")
	ui.Info("")
	ui.Info("Commands:")
	ui.Info("  add <name>       Add a new epic")
	ui.Info("  list             List all epics")
	ui.Info("  done <name>      Mark an epic as done")
	ui.Info("  remove <name>    Remove an epic")
}
