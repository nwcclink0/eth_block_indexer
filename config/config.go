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
log:
  format: "string" # string or json
  access_log: "stdout" # stdout: output to console,or define log path like "log/access_log"
  access_level: "debug"
  error_log: "stderr" # stderr: output to console,or define log path like "log/error_log"
  error_level: "error"
`)

type ConfYaml struct {
	Core SectionCore `yaml:"core"`
	Log  SectionLog  `yaml:"log"`
}

type SectionCore struct {
	StartBlockNum uint64 `yaml:"start_block_num:"`
	WorkerNum     int64  `yaml:"worker_num"`
	QueueNum      int64  `yaml:"queue_num"`
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
	fmt.Print(conf.Core)

	//Log
	conf.Log.Format = viper.GetString("log.format")
	conf.Log.AccessLog = viper.GetString("log.access_log")
	conf.Log.AccessLevel = viper.GetString("log.access_level")
	conf.Log.ErrorLog = viper.GetString("log.error_log")
	conf.Log.ErrorLevel = viper.GetString("log.error_level")

	return conf, nil
}
