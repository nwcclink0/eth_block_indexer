package service

import (
	"context"
	"fmt"
	"github.com/ackermanx/ethclient"
	"math/big"
	"time"
)

const binanceMainnet = `https://data-seed-prebsc-2-s3.binance.org:8545`

type ethBlockIndexer struct {
	LastScanBlockNum uint64
	ctx              context.Context
	cancel           context.CancelFunc
	dialContext      *ethclient.Client
}

func ShowBlockInfo(blockNum uint64) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	dialContext, err := ethclient.DialContext(ctx, binanceMainnet)
	block, err := dialContext.BlockByNumber(ctx, new(big.Int).SetUint64(blockNum))
	if err != nil {
		panic(err)
	}
	fmt.Println("================= block ================")
	fmt.Println("latest block num :", blockNum)
	fmt.Println("latest block hash: ", block.Hash())
	fmt.Println("latest block time: ", block.Time())
	fmt.Println("latest block parent hash: ", block.ParentHash())
	transations := block.Transactions()
	for i := 0; i < len(transations); i++ {
		trans := transations[i]
		fmt.Println("transation hash: ", trans.Hash())
		fmt.Println("transation from: ", trans.To())
		fmt.Println("transation to: ", trans.To())
		fmt.Println("transation nonce: ", trans.Nonce())
		fmt.Println("transation data: ", trans.Data())
		fmt.Println("transation value: ", trans.Value())
		receipt, err := dialContext.TransactionReceipt(ctx, trans.Hash())
		if err != nil {
			LogError.Error(err)
		}
		if receipt != nil {
			for j := 0; j < len(receipt.Logs); j++ {
				log := receipt.Logs[j]
				fmt.Println("log index:", log.Index)
				fmt.Println("log data", log.Data)
			}
		} else {
			if err != nil {
				fmt.Println(err)
			}
		}
	}
	fmt.Println("=================================")
	cancel()
}

func (indexer *ethBlockIndexer) Run() {
	//for {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	c, err := ethclient.DialContext(ctx, binanceMainnet)
	if err != nil {
		LogError.Error(err)
	}

	//get last block num
	lastBlockNumber, err := c.BlockNumber(ctx)
	if err != nil {
		LogError.Error(err)
	}
	cancel()

	//add new Block
	for blockNumber := indexer.LastScanBlockNum; blockNumber < lastBlockNumber; blockNumber++ {
		go Indexing(blockNumber)
		break
	}
	//}
}

type EthBlockIndexer interface {
	Run()
}

func NewIndexer(initBlockNumber uint64) EthBlockIndexer {
	indexer := &ethBlockIndexer{}
	indexer.LastScanBlockNum = initBlockNumber

	return indexer
}
