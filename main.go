/**
 *
 * @author nghiatc
 * @since Sep 30, 2020
 */

package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"ntc-gsuperwss/nwsc"
	"ntc-gsuperwss/nwss"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/congnghia0609/ntc-gconf/nconf"
	"github.com/natefinch/lumberjack"
)

// //// Declare Global
// // WSServer
var dpnwss *nwss.DPNWSServer

// var htnwss *nwss.HTNWSServer
// var csnwss *nwss.CSNWSServer
var tknwss *nwss.TKNWSServer

// var crnwss *nwss.CRNWSServer

// // WSClient
var dpnwsc *nwsc.NWSClient

// var csnwsc *nwsc.NWSClient
// var htnwsc *nwsc.NWSClient
var tknwsc *nwsc.NWSClient

// var crnwsc *nwsc.NWSClient
// var rsnwsc *nwsc.NWSClient

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

// increaseLimit increase resources limitations: ulimit -a
func increaseLimit() {
	var rlimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rlimit); err != nil {
		panic(err)
	}
	rlimit.Cur = rlimit.Max
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rlimit); err != nil {
		panic(err)
	}
	log.Printf("rlimit.Max = %d\n", rlimit.Max)
	log.Printf("rlimit.Cur = %d\n", rlimit.Cur)
}

// https://github.com/eranyanay/1m-go-websockets/
// https://github.com/gobwas/ws
func main() {
	// ////// -------------------- Init System -------------------- //////
	// Init NConf
	InitNConf()

	// //// init Logger
	// if "development" != nconf.GetEnv() {
	// 	log.Printf("============== LogFile: /data/log/ntc-gwss/ntc-gwss.log")
	// 	initLogger()
	// }

	// Increase resources limitations
	increaseLimit()

	// Enable pprof hooks
	go func() {
		if err := http.ListenAndServe("localhost:6060", nil); err != nil {
			log.Fatalf("pprof failed: %v", err)
		}
	}()

	//// initMapSymbol
	nwss.InitMapSymbol()

	////// -------------------- Start NWSServer -------------------- //////
	//// Run DPWSServer
	dpnwss = nwss.NewDPNWSServer(nwss.NameDPNWSS)
	log.Printf("======= DPNWSServer[%s] is ready...", dpnwss.GetName())
	go dpnwss.Start()

	//// Run TKNWSServer
	tknwss = nwss.NewTKNWSServer(nwss.NameTKNWSS)
	log.Printf("======= TKNWSServer[%s] is ready...", tknwss.GetName())
	go tknwss.Start()

	////// -------------------- Start NWSClient -------------------- //////
	// // DPNWSClient
	dpnwsc = nwsc.NewDPNWSClient()
	defer dpnwsc.Close()
	go dpnwsc.StartDPWSClient()

	// // // TKWSClient
	// tknwsc = nwsc.NewTKWSClient()
	// defer tknwsc.Close()
	// go tknwsc.StartTKWSClient()

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
