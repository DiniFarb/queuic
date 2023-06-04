package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/dinifarb/mlog"
	"github.com/dinifarb/queuic/pkg/server"
)

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
	switch strings.ToLower(logLevel) {
	case "debug":
		mlog.SetLevel(mlog.Ldebug)
	case "info":
		mlog.SetLevel(mlog.Linfo)
	case "warn":
		mlog.SetLevel(mlog.Lwarn)
	case "error":
		mlog.SetLevel(mlog.Lerror)
	default:
		mlog.Warn("LOG_LEVEL env variable is not set use default level")
		mlog.SetLevel(mlog.Linfo)
	}
	keyString := os.Getenv("QUEUEIC_KEY_STRING")
	if keyString == "" {
		mlog.Warn("QUEUEIC_KEY_STRING env variable is not set use default key")
		keyString = "QUEUEIC"
	}
	svr := server.NewQueuicServer(keyString)
	if err := svr.Serve(); err != nil {
		mlog.Error("server error: %v", err)
		os.Exit(1)
	}
}
