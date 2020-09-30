/**
 *
 * @author nghiatc
 * @since Aug 8, 2018
 */

package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"

	"github.com/congnghia0609/ntc-gconf/nconf"

	"github.com/natefinch/lumberjack"
)

// //// Declare Global
// // WSServer
// var dpwss *wss.DPWSServer
// var htwss *wss.HTWSServer
// var cswss *wss.CSWSServer
// var tkwss *wss.TKWSServer
// var crwss *wss.CRWSServer

// // WSClient
// var dpwsc *wsc.NWSClient
// var cswsc *wsc.NWSClient
// var htwsc *wsc.NWSClient
// var tkwsc *wsc.NWSClient
// var crwsc *wsc.NWSClient
// var rswsc *wsc.NWSClient

// InitNConf init file config
func InitNConf() {
	_, b, _, _ := runtime.Caller(0)
	wdir := filepath.Dir(b)
	fmt.Println("wdir:", wdir)
	nconf.Init(wdir)
}

// https://github.com/natefinch/lumberjack
func initLogger() {
	log.SetOutput(&lumberjack.Logger{
		Filename:   "/data/log/ntc-gsuperwss/ntc-gsuperwss.log",
		MaxSize:    10,   // 10 megabytes. Defaults to 100 MB.
		MaxBackups: 3,    // maximum number of old log files to retain.
		MaxAge:     28,   // maximum number of days to retain old log files
		Compress:   true, // disabled by default
	})
}

func main() {
	////// -------------------- Init System -------------------- //////
	// Init NConf
	InitNConf()

	//// init Logger
	if "development" != nconf.GetEnv() {
		log.Printf("============== LogFile: /data/log/ntc-gwss/ntc-gwss.log")
		initLogger()
	}

	////// -------------------- Start WSServer -------------------- //////
	// //// Run DPWSServer
	// dpwss = wss.NewDPWSServer(wss.NameDPWSS)
	// log.Printf("======= DPWSServer[%s] is ready...", dpwss.GetName())
	// go dpwss.Start()

	////// -------------------- Start WSClient -------------------- //////
	// // // DPWSClient
	// dpwsc = wsc.NewDPWSClient()
	// defer dpwsc.Close()
	// go dpwsc.StartDPWSClient()

	////// -------------------- Start WebServer -------------------- //////
	// // StartWebServer
	// go server.StartWebServer("webserver")

	// Hang thread Main.
	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C) SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)
	// Block until we receive our signal.
	<-c
	log.Println("################# End Main #################")
}
