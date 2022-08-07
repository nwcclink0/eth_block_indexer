package service

import (
	"context"
	"database/sql"
	"encoding/hex"
	"github.com/ackermanx/ethclient"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"math/big"
	"time"
)

type Block struct {
	gorm.Model
	BlockNum   uint64 `json:"block_num"`
	BlockHash  []byte `json:"block_hash"`
	BlockTime  uint64 `json:"block_time"`
	ParentHash []byte `json:"parent_hash"`
}

type BlockContainer struct {
	Blocks []Block `json:"blocks"`
}

type BlockWithTransactions struct {
	Block
	Transactions []string `json:"transactions"`
}

type BlockSummary struct {
	gorm.Model
	LastBlockNum uint64
}

type Transaction struct {
	gorm.Model
	TxHash   []byte `json:"tx_hash"`
	BlockNum uint64
	//TODO from need to add
	//From     common.Hash `json:"from"`
	To    []byte `json:"to"`
	Nonce uint64 `json:"nonce"`
	Data  []byte `json:"data"`
	Value uint64 `json:"value"`
}

type TransactionLog struct {
	gorm.Model
	TxHash []byte `json:"tx_hash"`
	Index  uint   `json:"index"`
	Data   []byte `json:"data"`
}

type TransactionWithLog struct {
	Transaction
	Logs []TransactionLog `json:"logs"`
}

const dsn = "host=localhost user=yt dbname=eth_block_index port=5432 sslmode=disable"

func Indexing(blockNum uint64) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		LogError.Error(err)
		panic(err)
	}

	err = db.AutoMigrate(&Block{})
	if err != nil {
		return
	}
	err = db.AutoMigrate(&BlockSummary{})
	if err != nil {
		return
	}
	err = db.AutoMigrate(&Transaction{})
	if err != nil {
		return
	}
	err = db.AutoMigrate(&TransactionLog{})
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	dialContext, err := ethclient.DialContext(ctx, binanceMainnet)
	block, err := dialContext.BlockByNumber(ctx, new(big.Int).SetUint64(blockNum))

	//check
	var blockInDb Block
	result := db.First(&blockInDb, Block{
		BlockNum: block.NumberU64(),
	})
	if result.Error == nil {
		//update
		db.Model(&blockInDb).Updates(Block{
			BlockNum:   block.NumberU64(),
			BlockHash:  block.Hash().Bytes(),
			BlockTime:  block.Time(),
			ParentHash: block.ParentHash().Bytes(),
		})
		transactions := block.Transactions()
		for i := 0; i < len(transactions); i++ {
			transaction := transactions[i]
			var dbTransaction Transaction

			result := db.First(&dbTransaction, Transaction{TxHash: transaction.Hash().Bytes()})
			if result.Error == nil {
				db.Create(&Transaction{
					BlockNum: block.NumberU64(),
					TxHash:   transaction.Hash().Bytes(),
					To:       transaction.To().Bytes(),
					Nonce:    transaction.Nonce(),
					Data:     transaction.Data(),
					Value:    transaction.Value().Uint64(),
				})
			} else {
				db.Model(&dbTransaction).Updates(
					&Transaction{
						BlockNum: block.NumberU64(),
						TxHash:   transaction.Hash().Bytes(),
						Nonce:    transaction.Nonce(),
						Data:     transaction.Data(),
						Value:    transaction.Value().Uint64(),
					})
			}
			//TODO add receipt update;
		}
	} else {
		// Insert block and related transactions

		db.Create(&Block{
			BlockNum:   block.NumberU64(),
			BlockHash:  block.Hash().Bytes(),
			BlockTime:  block.Time(),
			ParentHash: block.ParentHash().Bytes(),
		})

		db.Create(&BlockSummary{LastBlockNum: block.NumberU64()})

		transactions := block.Transactions()
		for i := 0; i < len(transactions); i++ {
			transaction := transactions[i]
			db.Create(&Transaction{
				BlockNum: block.NumberU64(),
				TxHash:   transaction.Hash().Bytes(),
				To:       transaction.To().Bytes(),
				Nonce:    transaction.Nonce(),
				Data:     transaction.Data(),
				Value:    transaction.Value().Uint64(),
			})

			receipt, err := dialContext.TransactionReceipt(ctx, transaction.Hash())
			if err != nil {
				LogError.Error(err)
			}
			if receipt != nil {
				for j := 0; j < len(receipt.Logs); j++ {
					log := receipt.Logs[j]
					db.Create(&TransactionLog{
						TxHash: transaction.Hash().Bytes(),
						Index:  log.Index,
						Data:   log.Data,
					})
				}
			}
		}
	}
	cancel()
}

func GetLastNBlocks(n uint64) *BlockContainer {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	var blockContainer BlockContainer
	if err != nil {
		LogError.Error(err)
		return &blockContainer
	}

	var blockSummary BlockSummary
	result := db.First(&blockSummary)
	if result.Error == nil {
		startBlockNum := blockSummary.LastBlockNum - n + 1
		for ; startBlockNum <= blockSummary.LastBlockNum; startBlockNum++ {
			var block Block
			result := db.First(&block, Block{BlockNum: startBlockNum})
			if result.Error != nil {
				LogAccess.Debug("block number:", startBlockNum, " didn't exist in db")
			} else {
				LogAccess.Debug(" block number hash: ", string(block.BlockHash))
				blockContainer.Blocks = append(blockContainer.Blocks, block)
			}
		}
		return &blockContainer
	} else {
		return &blockContainer
	}
}

func GetBlockById(blockNum uint64) *BlockWithTransactions {
	var blockWithTransaction BlockWithTransactions
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		LogError.Error(err)
		return &blockWithTransaction
	}
	var block Block
	result := db.First(&block, Block{
		BlockNum: blockNum,
	})
	if result.Error == nil {
		blockWithTransaction.BlockNum = block.BlockNum
		blockWithTransaction.BlockHash = block.BlockHash
		blockWithTransaction.BlockTime = block.BlockTime
		blockWithTransaction.ParentHash = block.ParentHash

		var transaction []Transaction
		result := db.Find(&transaction, Transaction{BlockNum: blockNum})
		rows, err := result.Rows()
		if err != nil {
			LogAccess.Debug(err)
		}
		defer func(rows *sql.Rows) {
			err := rows.Close()
			if err != nil {
				LogAccess.Debug(err)
			}
		}(rows)
		for rows.Next() {
			var transaction Transaction
			err := db.ScanRows(rows, &transaction)
			if err != nil {
				LogAccess.Debug(err)
			}
			blockWithTransaction.Transactions = append(blockWithTransaction.Transactions, hex.EncodeToString(transaction.TxHash))
		}
		return &blockWithTransaction
	} else {
		return &blockWithTransaction
	}
}

func getTransactionByTxHash(txHashStr string) *TransactionWithLog {
	var transactionWithLog TransactionWithLog
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		LogError.Error(err)
		return &transactionWithLog
	}
	var transaction Transaction
	txHash, err := hex.DecodeString(txHashStr)
	if err != nil {
		LogError.Error(err)
		return &transactionWithLog
	}
	result := db.First(&transaction, Transaction{TxHash: txHash})
	if result.Error == nil {
		transactionWithLog.TxHash = transaction.TxHash
		transactionWithLog.To = transaction.To
		transactionWithLog.Nonce = transaction.Nonce
		transactionWithLog.Data = transaction.Data
		transactionWithLog.Value = transaction.Value

		var transactionLogs []TransactionLog
		result := db.Find(&transactionLogs, TransactionLog{TxHash: transaction.TxHash})
		if result.Error != nil {
			LogAccess.Debug("didn't exist log of transaction tx_hash:", string(transaction.TxHash))
			return &transactionWithLog
		} else {
			rows, err := result.Rows()
			if err != nil {
				return &transactionWithLog
			}
			defer func(rows *sql.Rows) {
				err := rows.Close()
				if err != nil {
					LogAccess.Debug(err)
				}
			}(rows)
			for rows.Next() {
				var transactionLog TransactionLog
				err := db.ScanRows(rows, &transactionLog)
				if err != nil {
					LogAccess.Debug(err)
				}
				transactionWithLog.Logs = append(transactionWithLog.Logs, transactionLog)
			}
		}
		return &transactionWithLog
	} else {
		return &transactionWithLog
	}
}
