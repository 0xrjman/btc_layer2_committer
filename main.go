package main

import (
	"github.com/mapprotocol/btc_layer2_committer/config"
	"github.com/mapprotocol/btc_layer2_committer/task"
)

func main() {
	//alarm.ValidateEnv()

	config.Init()

	task.Run()
}
