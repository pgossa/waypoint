package job

import (
	"os"

	"waypoint/model"
	"waypoint/storage"
	"waypoint/ui"
)

func jobAdd(args []string) {
	if len(args) != 1 {
		jobAddUsage()
		os.Exit(1)
	}
	job, err := model.CreateJob(args[0])
	if err != nil {
		ui.Error(err.Error())
		jobUsage()
		os.Exit(1)
	}
	if err = storage.SaveJob(job); err != nil {
		ui.Error(err.Error())
		jobUsage()
		os.Exit(1)
	}
	ui.Success("Job Added !")
	os.Exit(0)
}

func jobAddUsage() {
	ui.Info("Usage: wpt job add <name>")
	ui.Info("  <name>  Name of the job (max 64 chars, alphanumeric and ._- allowed)")
	ui.Info("")
	ui.Info("Example:")
	ui.Info("  wpt job add \"fix something\"")
}
