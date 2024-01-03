package config

import (
	"fmt"
	"github.com/mapprotocol/btc_layer2_committer/utils"
	"math/big"
	"testing"
)

func Test_genCfg(t *testing.T) {
	configFile = "./config.json"
	config0 := &Config{}

	config0.TestNet = true
	config0.AtlasURL = "https://rpc.maplabs.io"
	config0.Sender = "ac17879d2966723ea4c67eda40ff21f63856740aade87a8cce4e764a4bc1b2a1"
	config0.LatestCheckPoint = &utils.CheckPoint{
		Height: big.NewInt(5000000),
		Root:   "0xa447fbc19970f936ba22d20cce4996bbdc690253e813141e13ba7ccc55cfc137",
	}
	CfgParams = config0
	err := SaveConfig()
	if err != nil {
		fmt.Println("Error saving configuration:", err)
		return
	}

	fmt.Println("Configuration saved successfully.")
}
func Test_01(t *testing.T) {
	configFile = "./config.json"
	config, err := loadConfig("config.json")
	if err != nil {
		fmt.Println("Error loading configuration:", err)
		return
	}
	config.TestNet = true
	config.AtlasURL = "https://rpc.maplabs.io"
	config.Sender = "a8d7879d2966723ea4c67eda40ff21f63856740aade87a8cce4e764a4bc16ea1"
	config.LatestCheckPoint = &utils.CheckPoint{
		Height: big.NewInt(5000000),
		Root:   "0xa447fbc19970f936ba22d20cce4996bbdc690253e813141e13ba7ccc55cfc137",
	}
	err = SaveConfig()
	if err != nil {
		fmt.Println("Error saving configuration:", err)
		return
	}

	fmt.Println("Configuration saved successfully.")
}
