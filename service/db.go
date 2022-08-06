package service

import (
	"context"
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

func Indexing(blockNum uint64) {
	dsn := "host=localhost user=yt dbname=eth_block_index port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		LogError.Error(err)
		panic(err)
	}

	err = db.AutoMigrate(&Block{})
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
						To:       transaction.To().Bytes(),
						Nonce:    transaction.Nonce(),
						Data:     transaction.Data(),
						Value:    transaction.Value().Uint64(),
					})
			}
			//TODO add receipt update;
		}
	} else {
		// Insert
		db.Create(&Block{
			BlockNum:   block.NumberU64(),
			BlockHash:  block.Hash().Bytes(),
			BlockTime:  block.Time(),
			ParentHash: block.ParentHash().Bytes(),
		})
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
