package main

import (
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/btc_layer2_committer/config"
	"github.com/mapprotocol/btc_layer2_committer/task"
	"os"
)

func main() {
	//alarm.ValidateEnv()
	log.SetDefault(log.NewLogger(log.NewTerminalHandler(os.Stderr, true)))

	config.Init()
	task.Run()
}
