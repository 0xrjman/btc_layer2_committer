package main

import (
	"github.com/mapprotocol/atlas_committer/config"
	"github.com/mapprotocol/atlas_committer/task"
	"github.com/mapprotocol/atlas_committer/utils/alarm"
)

func main() {
	alarm.ValidateEnv()

	config.Init()

	task.Run()
}
