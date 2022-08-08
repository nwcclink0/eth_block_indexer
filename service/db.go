package service

import (
	"context"
	"database/sql"
	"encoding/hex"
	"github.com/ackermanx/ethclient"
	"github.com/ethereum/go-ethereum/core/types"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"math/big"
	"time"
)

type Block struct {
	gorm.Model
	BlockNum   uint64
	BlockHash  []byte
	BlockTime  uint64
	ParentHash []byte
}

type BlockJSN struct {
	BlockNum   uint64 `json:"block_num"`
	BlockHash  string `json:"block_hash"`
	BlockTime  uint64 `json:"block_time"`
	ParentHash string `json:"parent_hash"`
}

type BlockContainerJSN struct {
	Blocks []BlockJSN `json:"blocks"`
}

type BlockWithTransactionsJSN struct {
	BlockJSN
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
	From     []byte `json:"from"`
	To       []byte `json:"to"`
	Nonce    uint64 `json:"nonce"`
	Data     []byte `json:"data"`
	Value    uint64 `json:"value"`
}

type TransactionJSN struct {
	TxHash string `json:"tx_hash"`
	From   string `json:"from"`
	To     string `json:"to"`
	Nonce  uint64 `json:"nonce"`
	Data   string `json:"data"`
	Value  uint64 `json:"value"`
}

type TransactionLog struct {
	gorm.Model
	TxHash []byte
	Index  uint
	Data   []byte
}

type TransactionLogJSN struct {
	Index uint   `json:"index"`
	Data  string `json:"data"`
}

type TransactionWithLogJSN struct {
	TransactionJSN
	Logs []TransactionLogJSN `json:"logs"`
}

const dsn = "host=localhost user=yt dbname=eth_block_index port=5432 sslmode=disable"

func InitDb() {
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		LogError.Error(err)
		panic(err)
	}
}

func Indexing(blockNum uint64) {
	err := db.AutoMigrate(&Block{})
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
	cancel()

	//check block existence
	var blockInDb Block
	result := db.First(&blockInDb, Block{
		BlockNum: block.NumberU64(),
	})
	if result.Error == nil {
		//update block
		db.Model(&blockInDb).Updates(Block{
			BlockNum:   block.NumberU64(),
			BlockHash:  block.Hash().Bytes(),
			BlockTime:  block.Time(),
			ParentHash: block.ParentHash().Bytes(),
		})
		transactions := block.Transactions()
		for i := 0; i < len(transactions); i++ {
			// add or update transaction
			transaction := transactions[i]
			if transaction == nil {
				continue
			}
			var to = make([]byte, 0)
			if transaction.To() != nil {
				to = transaction.To().Bytes()
			} else {
				LogAccess.Debug("transaction to is null")
			}
			var dbTransaction Transaction
			chainId, err := dialContext.NetworkID(context.Background())
			if err != nil {
				LogError.Error(err)
			}
			msg, err := transaction.AsMessage(types.NewEIP155Signer(chainId), block.BaseFee())
			if err != nil {
				LogError.Error(err)
			}
			result := db.First(&dbTransaction, Transaction{TxHash: transaction.Hash().Bytes()})
			if result.Error == nil {
				db.Create(&Transaction{
					BlockNum: block.NumberU64(),
					TxHash:   transaction.Hash().Bytes(),
					From:     msg.From().Hash().Bytes(),
					To:       to,
					Nonce:    transaction.Nonce(),
					Data:     transaction.Data(),
					Value:    transaction.Value().Uint64(),
				})
			} else {
				db.Model(&dbTransaction).Updates(
					&Transaction{
						BlockNum: block.NumberU64(),
						TxHash:   transaction.Hash().Bytes(),
						From:     msg.From().Hash().Bytes(),
						To:       to,
						Nonce:    transaction.Nonce(),
						Data:     transaction.Data(),
						Value:    transaction.Value().Uint64(),
					})
			}

			// add new log for a transaction
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			dialContext, err := ethclient.DialContext(ctx, binanceMainnet)
			receipt, err := dialContext.TransactionReceipt(ctx, transaction.Hash())
			cancel()
			if err != nil {
				LogError.Error(err)
			}
			if receipt != nil {
				for j := 0; j < len(receipt.Logs); j++ {
					log := receipt.Logs[j]
					var dbTransactionLog TransactionLog
					result := db.First(&dbTransactionLog, TransactionLog{TxHash: transaction.Hash().Bytes(), Index: log.Index})
					if result.Error == nil {
						continue
					}
					db.Create(&TransactionLog{
						TxHash: transaction.Hash().Bytes(),
						Index:  log.Index,
						Data:   log.Data,
					})
				}
			}

		}
	} else {
		// Insert block and related transactions

		db.Create(&Block{
			BlockNum:   block.NumberU64(),
			BlockHash:  block.Hash().Bytes(),
			BlockTime:  block.Time(),
			ParentHash: block.ParentHash().Bytes(),
		})

		var blockSummary BlockSummary
		result := db.First(&blockSummary)
		if result.Error != nil {
			db.Create(&BlockSummary{LastBlockNum: block.NumberU64()})
		} else {
			db.Model(&blockSummary).Updates(&BlockSummary{LastBlockNum: block.NumberU64()})
		}

		transactions := block.Transactions()
		for i := 0; i < len(transactions); i++ {
			transaction := transactions[i]
			if transaction == nil {
				continue
			}
			chainId, err := dialContext.NetworkID(context.Background())
			if err != nil {
				LogError.Error(err)
			}
			msg, err := transaction.AsMessage(types.NewEIP155Signer(chainId), block.BaseFee())
			if err != nil {
				LogError.Error(err)
			}
			var to = make([]byte, 0)
			if transaction.To() != nil {
				to = transaction.To().Bytes()
			} else {
				LogAccess.Debug("transaction to is null")
			}
			db.Create(&Transaction{
				BlockNum: block.NumberU64(),
				TxHash:   transaction.Hash().Bytes(),
				From:     msg.From().Hash().Bytes(),
				To:       to,
				Nonce:    transaction.Nonce(),
				Data:     transaction.Data(),
				Value:    transaction.Value().Uint64(),
			})
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			dialContext, err := ethclient.DialContext(ctx, binanceMainnet)
			receipt, err := dialContext.TransactionReceipt(ctx, transaction.Hash())
			cancel()

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

func hashBytesToStringWithPrefix(hash []byte) string {
	return "0x" + hex.EncodeToString(hash)
}

func GetLastNBlocks(n uint64) *BlockContainerJSN {
	var blockContainer BlockContainerJSN
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
				blockJSN := BlockJSN{
					BlockNum:   block.BlockNum,
					BlockHash:  hashBytesToStringWithPrefix(block.BlockHash),
					BlockTime:  block.BlockTime,
					ParentHash: hashBytesToStringWithPrefix(block.ParentHash),
				}
				blockContainer.Blocks = append(blockContainer.Blocks, blockJSN)
			}
		}
		return &blockContainer
	} else {
		return &blockContainer
	}
}

// GetBlockById block id defined as block number
func GetBlockById(blockNum uint64) *BlockWithTransactionsJSN {
	var blockWithTransactionsJSN BlockWithTransactionsJSN
	var block Block
	result := db.First(&block, Block{
		BlockNum: blockNum,
	})
	if result.Error == nil {
		blockWithTransactionsJSN.BlockNum = block.BlockNum
		blockWithTransactionsJSN.BlockHash = hashBytesToStringWithPrefix(block.BlockHash)
		blockWithTransactionsJSN.BlockTime = block.BlockTime
		blockWithTransactionsJSN.ParentHash = hashBytesToStringWithPrefix(block.ParentHash)

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
			blockWithTransactionsJSN.Transactions =
				append(blockWithTransactionsJSN.Transactions, hashBytesToStringWithPrefix(transaction.TxHash))
		}
		return &blockWithTransactionsJSN
	} else {
		return &blockWithTransactionsJSN
	}
}

func getTransactionByTxHash(txHashWithPrefixStr string) *TransactionWithLogJSN {
	var transactionWithLogJSN TransactionWithLogJSN
	var transaction Transaction
	prefix := txHashWithPrefixStr[0:2]
	if prefix != "0x" {
		LogAccess.Debug("incorrect txHash: ", txHashWithPrefixStr)
		return &transactionWithLogJSN
	}
	txHashStr := txHashWithPrefixStr[2:]
	txHash, err := hex.DecodeString(txHashStr)
	if err != nil {
		LogError.Error(err)
		return &transactionWithLogJSN
	}
	result := db.First(&transaction, Transaction{TxHash: txHash})
	if result.Error == nil {
		transactionWithLogJSN.TxHash = hashBytesToStringWithPrefix(transaction.TxHash)
		transactionWithLogJSN.From = hashBytesToStringWithPrefix(transaction.From)
		transactionWithLogJSN.To = hashBytesToStringWithPrefix(transaction.To)
		transactionWithLogJSN.Nonce = transaction.Nonce
		transactionWithLogJSN.Data = hashBytesToStringWithPrefix(transaction.Data)
		transactionWithLogJSN.Value = transaction.Value

		var logs = make([]TransactionLogJSN, 0)
		transactionWithLogJSN.Logs = logs

		var transactionLogs []TransactionLog
		result := db.Find(&transactionLogs, TransactionLog{TxHash: transaction.TxHash})
		if result.Error != nil {
			LogAccess.Debug("didn't exist log of transaction tx_hash:", string(transaction.TxHash))
			return &transactionWithLogJSN
		} else {
			rows, err := result.Rows()
			if err != nil {
				return &transactionWithLogJSN
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
				transactionLogJSN := TransactionLogJSN{
					Index: transactionLog.Index,
					Data:  hashBytesToStringWithPrefix(transactionLog.Data),
				}
				transactionWithLogJSN.Logs = append(transactionWithLogJSN.Logs, transactionLogJSN)
			}
		}
		return &transactionWithLogJSN
	} else {
		return &transactionWithLogJSN
	}
}
