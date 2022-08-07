package service

import "strconv"

func InitWorker(workerNum int64, queueNum int64) {
	LogAccess.Debug("worker number is " + strconv.FormatInt(workerNum,
		10) + ", " +
		"queue number is " + strconv.FormatInt(queueNum, 10))
	QueueIndexingBlockNum = make(chan uint64, queueNum)
	for i := int64(0); i < workerNum; i++ {
		go startWorker()
	}
}

func startWorker() {
	for {
		blockNum := <-QueueIndexingBlockNum
		LogAccess.Debug("indexing block number: ", blockNum)
		Indexing(blockNum)
	}
}
