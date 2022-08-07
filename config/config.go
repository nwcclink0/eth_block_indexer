package config

import (
	"bytes"
	"fmt"
	"github.com/spf13/viper"
	"io/ioutil"
	"strings"
)

var defaultConf = []byte(`
core:
  start_block_num: 21709284
  worker_num: 0 # default worker number is runtime.NumCPU()
  queue_num: 0 # default queue number is 8192
  address: ""
  http_port: "8080"
  https_port: "8081"
  mode: "release"
api:
  blocks_uri: "/blocks"
  block_by_id_uri: "/blocks/:id"
  transaction_uri: "/transaction/:txHash"
log:
  format: "string" # string or json
  access_log: "stdout" # stdout: output to console,or define log path like "log/access_log"
  access_level: "debug"
  error_log: "stderr" # stderr: output to console,or define log path like "log/error_log"
  error_level: "error"
`)

type ConfYaml struct {
	Core SectionCore `yaml:"core"`
	API  SectionAPI  `yaml:"api"`
	Log  SectionLog  `yaml:"log"`
}

type SectionCore struct {
	StartBlockNum uint64 `yaml:"start_block_num:"`
	WorkerNum     int64  `yaml:"worker_num"`
	QueueNum      int64  `yaml:"queue_num"`
	Address       string `yaml:"address"`
	HttpPort      string `yaml:"http_port"`
	HttpsPort     string `yaml:"https_port"`
	Mode          string `yaml:"mode"`
}

type SectionAPI struct {
	BlocksURI      string `yaml:"blocks_uri"`
	BlockByIdURI   string `yaml:"block_by_id_uri"`
	TransactionURI string `yaml:"transaction_uri"`
}

type SectionLog struct {
	Format      string `yaml:"format"`
	AccessLog   string `yaml:"access_log"`
	AccessLevel string `yaml:"access_level"`
	ErrorLog    string `yaml:"error_log"`
	ErrorLevel  string `yaml:"error_level"`
}

func LoadConf(confPath string) (ConfYaml, error) {
	var conf ConfYaml

	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	viper.SetEnvPrefix("eth_block_indexer")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if confPath != "" {
		content, err := ioutil.ReadFile(confPath)

		if err != nil {
			return conf, err
		}

		if err := viper.ReadConfig(bytes.NewBuffer(content)); err != nil {
			return conf, err
		}
	} else {
		viper.AddConfigPath("/etc/eth_block_indexer")
		viper.SetConfigName("config")

		if err := viper.ReadInConfig(); err == nil {
			fmt.Println("Using config file:", viper.ConfigFileUsed())
		} else {
			if err := viper.ReadConfig(bytes.NewBuffer(
				defaultConf)); err != nil {
				return conf, err
			}
		}
	}
	//Core
	conf.Core.StartBlockNum = uint64(viper.GetInt("core.start_block_num"))
	conf.Core.WorkerNum = int64(viper.GetInt("core.worker_num"))
	conf.Core.QueueNum = int64(viper.GetInt("core.queue_num"))
	conf.Core.Address = viper.GetString("core.address")
	conf.Core.HttpPort = viper.GetString("core.http_port")
	conf.Core.HttpsPort = viper.GetString("core.https_port")
	conf.Core.Mode = viper.GetString("core.mode")
	fmt.Print(conf.Core)

	//API
	conf.API.BlocksURI = viper.GetString("api.blocks_uri")
	conf.API.BlockByIdURI = viper.GetString("api.block_by_id_uri")
	conf.API.TransactionURI = viper.GetString("api.transaction_uri")

	//Log
	conf.Log.Format = viper.GetString("log.format")
	conf.Log.AccessLog = viper.GetString("log.access_log")
	conf.Log.AccessLevel = viper.GetString("log.access_level")
	conf.Log.ErrorLog = viper.GetString("log.error_log")
	conf.Log.ErrorLevel = viper.GetString("log.error_level")

	return conf, nil
}
