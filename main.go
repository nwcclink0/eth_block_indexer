package main

import (
	"eth_block_indexer/config"
	"eth_block_indexer/service"
	"flag"
	"fmt"
	"github.com/coreos/go-systemd/daemon"
	"golang.org/x/sync/errgroup"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var (
		configFile string
	)

	flag.StringVar(&configFile, "c", "", "Configuration file path")
	flag.Usage = usage
	flag.Parse()

	var err error

	service.EthBlockIndexerConf, err = config.LoadConf(configFile)
	if err != nil {
		log.Fatalf("Load yaml config file error: '%v'", err)
		return
	}

	if err = service.InitLog(); err != nil {
		log.Fatalf("Can't load log module, error: %v", err)
	}

	service.LogError.Debug("loaded config, ", service.EthBlockIndexerConf)

	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGSEGV,
	)
	go func() {
		sig := <-signalChannel
		switch sig {
		case os.Interrupt:
			service.LogAccess.Debug("receive interrupt")
			os.Exit(0)
		case syscall.SIGTERM:
			service.LogAccess.Debug("receive sigterm")
			os.Exit(0)
		case syscall.SIGHUP:
			service.LogAccess.Debug("receive sighup")
		case syscall.SIGINT:
			service.LogAccess.Debug("receive sigint")
			os.Exit(0)
		case syscall.SIGQUIT:
			service.LogAccess.Debug("receive sigint")
			os.Exit(0)
		case syscall.SIGSEGV:
			service.LogAccess.Debug("receive sigsegv")
			os.Exit(0)
		}
	}()
	notify, err := daemon.SdNotify(false, daemon.SdNotifyReady)
	if notify == false {
		service.LogError.Debug("notify do not support")
	}
	if err != nil {
		service.LogError.Fatal(err)
	}
	service.InitWorker(service.EthBlockIndexerConf.Core.WorkerNum,
		service.EthBlockIndexerConf.Core.QueueNum)
	service.InitDb()
	indexer := service.NewIndexer(service.EthBlockIndexerConf.Core.StartBlockNum)

	var g errgroup.Group
	g.Go(service.RunHTTPServer)
	indexer.Run()
	if err = g.Wait(); err != nil {
		service.LogError.Fatal(err)
	}
}

var usageStr = `
Usage: [options]
Server Options:
	-c, --config <file>
`

func usage() {
	fmt.Printf("%s\n", usageStr)
	os.Exit(0)
}
