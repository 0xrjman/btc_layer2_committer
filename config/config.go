package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/mapprotocol/btc_layer2_committer/utils"
	"io/ioutil"
)

var (
	configFile string
)

type Config struct {
	TestNet          bool              `json:"testnet"`
	Sender           string            `json:"sender"`
	AtlasURL         string            `json:"atlas_url"`
	LatestCheckPoint *utils.CheckPoint `json:"latest_check_point"`
}

func loadConfig(filename string) (*Config, error) {
	var config Config

	fileContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(fileContent, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func SaveConfig() error {
	if CfgParams != nil {
		configJSON, err := json.MarshalIndent(*CfgParams, "", "    ")
		if err != nil {
			return err
		}

		err = ioutil.WriteFile(configFile, configJSON, 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

func Init() {
	flag.StringVar(&configFile, "config", "config/config.json", "path of config file")
	flag.Parse()
	if configFile == "" {
		flag.Usage()
	}
	cfg, err := loadConfig(configFile)
	if err != nil {
		errStr := fmt.Sprintf("viper read config is failed, err is %v configFile is %s ", err, configFile)
		panic(errStr)
	}

	CfgParams = cfg
}

var CfgParams *Config = nil
