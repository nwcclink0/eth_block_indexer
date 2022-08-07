package service

import (
	"eth_block_indexer/config"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	EthBlockIndexerConf   config.ConfYaml
	QueueIndexingBlockNum chan uint64
	LogAccess             *logrus.Logger
	LogError              *logrus.Logger
	db                    *gorm.DB
)
