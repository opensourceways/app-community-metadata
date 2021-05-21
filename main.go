package main

import (
	"github.com/gookit/color"
	"github.com/opensourceways/app-community-metadata/app"
	"github.com/opensourceways/app-community-metadata/application"
	"github.com/opensourceways/app-community-metadata/application/gitsync"
	_ "github.com/opensourceways/app-community-metadata/application/gitsync/plugins"
	"github.com/opensourceways/app-community-metadata/cache"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	Manager *gitsync.SyncManager
)

func init() {
	app.Bootstrap("./config")
	cache.InitCache()
	application.InitServer()
}
func main() {
	listenSignals()
	//init manager
	Manager, err := gitsync.NewSyncManager(application.Server().Group("/v1/metadata"))
	if err != nil {
		color.Error.Printf("failed to initialize sync manager %v\n", err)
		os.Exit(1)
	}
	err = Manager.Initialize()
	if err != nil {
		color.Error.Printf("failed to start manager %v\n ", err)
		os.Exit(1)
	}
	Manager.StartLoop()
	// init services
	color.Info.Printf("============  Begin Running(PID: %d) ============\n", os.Getpid())
	application.Run()
}

// listenSignals Graceful start/stop server
func listenSignals() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go handleSignals(sigChan)
}

// handleSignals handle process signal
func handleSignals(c chan os.Signal) {
	color.Info.Printf("Notice: System signal monitoring is enabled(watch: SIGINT,SIGTERM,SIGQUIT)\n")

	switch <-c {
	case syscall.SIGINT:
		color.Info.Printf("\nShutdown by Ctrl+C")
	case syscall.SIGTERM: // by kill
		color.Info.Printf("\nShutdown quickly")
	case syscall.SIGQUIT:
		color.Info.Printf("\nShutdown gracefully")
		// do graceful shutdown
	}

	// sync logs
	_ = app.Logger.Sync()
	_ = cache.Close()
	if Manager != nil {
		Manager.Close()
	}
	//sleep and exit
	time.Sleep(1e9 / 2)
	color.Info.Println("\nGoodBye...")

	os.Exit(0)
}
