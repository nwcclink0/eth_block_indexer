package service

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/ackermanx/ethclient"
	"github.com/facebookgo/grace/gracehttp"
	"github.com/gin-gonic/gin"
	"math/big"
	"net/http"
	"strconv"
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

	LogAccess.Debug("total scan blocks number: ", lastBlockNumber-indexer.LastScanBlockNum)

	//add new Block
	for blockNumber := indexer.LastScanBlockNum + 1; blockNumber <= lastBlockNumber; blockNumber++ {
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

func RunHTTPServer() (err error) {
	//config := &tls.Config{
	//	MinVersion: tls.VersionTLS12,
	//}
	//err = gracehttp.Serve(tlsServer(config), httpServer())
	err = gracehttp.Serve(httpServer())
	return
}

func tlsServer(config *tls.Config) *http.Server {
	return &http.Server{
		Addr:      EthBlockIndexerConf.Core.Address + ":" + EthBlockIndexerConf.Core.HttpsPort,
		TLSConfig: config,
		Handler:   RouterEngine(),
	}
}

func httpServer() *http.Server {
	return &http.Server{
		Addr:    EthBlockIndexerConf.Core.Address + ":" + EthBlockIndexerConf.Core.HttpPort,
		Handler: RouterEngine(),
	}
}

func RouterEngine() *gin.Engine {
	gin.SetMode(EthBlockIndexerConf.Core.Mode)

	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(LogMiddleware())

	router.GET(EthBlockIndexerConf.API.BlocksURI, queryBlocksHandler)
	router.GET(EthBlockIndexerConf.API.BlockByIdURI, queryBlockByIdHandler)
	router.GET(EthBlockIndexerConf.API.TransactionURI, queryTransactionHandler)
	router.GET("/welcome", rootHandler)

	return router
}

func rootHandler(context *gin.Context) {
	context.JSON(http.StatusOK, gin.H{
		"text": "Welcome to eth block indexer service",
	})
}

func queryBlocksHandler(context *gin.Context) {
	lastNBlockStr := context.Query("limit")
	lastNBlock, err := strconv.Atoi(lastNBlockStr)
	var emptyBlocks = make([]BlockJSN, 0)
	if err != nil {
		LogAccess.Debug("didn't contain last ", lastNBlockStr, " block")
		context.JSON(http.StatusOK, gin.H{
			"blocks": &emptyBlocks,
		})
		return
	}
	lastNBlockU64 := uint64(lastNBlock)
	blockContainer := GetLastNBlocks(lastNBlockU64)
	if blockContainer == nil {
		LogAccess.Debug("didn't contain last ", lastNBlock, " block")
		context.JSON(http.StatusOK, gin.H{
			"blocks": &emptyBlocks,
		})
		return
	} else {
		context.JSON(http.StatusOK, blockContainer)
	}
}

func queryBlockByIdHandler(context *gin.Context) {
	blockIdStr := context.Param("id")
	blockId, err := strconv.Atoi(blockIdStr)

	var emptyBlockWithTransactionsJSN BlockWithTransactionsJSN
	if err != nil {
		context.JSON(http.StatusOK, emptyBlockWithTransactionsJSN)
		return
	}
	blockIdkU64 := uint64(blockId)
	blockWithTransactions := GetBlockById(blockIdkU64)
	context.JSON(http.StatusOK, blockWithTransactions)
}

func queryTransactionHandler(context *gin.Context) {
	txHash := context.Param("txHash")
	if len(txHash) == 0 {
		transactionWithLog := TransactionWithLogJSN{}
		context.JSON(http.StatusOK, transactionWithLog)
	} else {
		transactionWithLog := getTransactionByTxHash(txHash)
		context.JSON(http.StatusOK, transactionWithLog)
	}
}
