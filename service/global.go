package service

import (
	"eth_block_indexer/config"
	"github.com/sirupsen/logrus"
)

var (
	EthBlockIndexerConf config.ConfYaml
	LogAccess           *logrus.Logger
	LogError            *logrus.Logger
)
