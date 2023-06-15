package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/dinifarb/mlog"
	"github.com/dinifarb/queuic/pkg/server"
)

var srv *server.QueuicServer

func main() {
	fmt.Println(`
     ___                   _      
    / _ \ _   _  ___ _   _(_) ___ 
   | | | | | | |/ _ \ | | | |/ __|
   | |_| | |_| |  __/ |_| | | (__ 
    \__\_\\__,_|\___|\__,_|_|\___|							  
   `)
	mlog.SetAppName("QUEUEIC")
	logLevel := os.Getenv("LOG_LEVEL")
	if strings.ToLower(logLevel) == "debug" {
		mlog.SetLevel(mlog.Ldebug)
	}
	keyString := os.Getenv("QUEUEIC_KEY_STRING")
	if keyString == "" {
		mlog.Warn("QUEUEIC_KEY_STRING env variable is not set use default key")
		keyString = "QUEUEIC"
	}
	srv = server.NewQueuicServer(keyString)
	if err := srv.LoadQueuesFromDisk(); err != nil {
		mlog.Error("failed to load queues from disk: %v", err)
		os.Exit(1)
	}
	manager := NewManager()
	go func() {
		if err := manager.Start(); err != nil {
			mlog.Error("failed to start manager: %v", err)
			os.Exit(1)
		}
	}()
	if err := srv.Serve(); err != nil {
		mlog.Error("server error: %v", err)
		os.Exit(1)
	}
}
