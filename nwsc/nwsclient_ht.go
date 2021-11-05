/**
 *
 * @author nghiatc
 * @since Aug 8, 2018
 */

package nwsc

import (
	"fmt"
	"github.com/congnghia0609/ntc-gsuperwss/nwss"
	"github.com/congnghia0609/ntc-gsuperwss/util"
	"log"
	"time"

	"github.com/congnghia0609/ntc-gconf/nconf"
	"github.com/gorilla/websocket"
)

func (nwsc *NWSClient) recvHT() {
	util.TCF{
		Try: func() {
			defer nwsc.Close()
			defer close(nwsc.done)
			for {
				_, message, err := nwsc.conn.ReadMessage()
				if err != nil {
					log.Println("read:", err)
					nwsc.Reconnect()
					// return
				}
				log.Printf("recvHT: %s", message)
				if len(message) > 0 {
					// HTNWSServer
					htnwss := nwss.GetInstanceHT(nwss.NameHTNWSS)
					if htnwss != nil {
						htnwss.BroadcastMsgByte(message)
					}
				}
			}
		},
		Catch: func(e util.Exception) {
			log.Printf("nwsc.recvHT Caught %v\n", e)
		},
		Finally: func() {
			//log.Println("Finally...")
		},
	}.Do()
}

func (nwsc *NWSClient) sendHT() {
	util.TCF{
		Try: func() {
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()

			for {
				select {
				case t := <-ticker.C:
					//err := nws.conn.WriteMessage(websocket.TextMessage, []byte(t.String()))
					msec := t.UnixNano() / 1000000
					///// 1. Historytrade Data.
					data := `{"p":"0.05567000","q":"1.84100000","c":1533886283334,"s":"ETH_BTC","t":` + fmt.Sprint(msec) + `,"e":"history_trade","k":514102,"m":true}`
					err := nwsc.conn.WriteMessage(websocket.TextMessage, []byte(data))
					if err != nil {
						log.Println("write:", err)
						//return
					}
				case <-nwsc.interrupt:
					log.Println("interrupt")
					// To cleanly close a connection, a client should send a close
					// frame and wait for the server to close the connection.
					err := nwsc.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
					if err != nil {
						log.Println("write close:", err)
						return
					}
					select {
					case <-nwsc.done:
					case <-time.After(time.Second):
					}
					nwsc.Close()
					return
				}
			}
		},
		Catch: func(e util.Exception) {
			log.Printf("nwsc.sendHT Caught %v\n", e)
		},
		Finally: func() {
			//log.Println("Finally...")
		},
	}.Do()
}

// NewHTNWSClient new instance of NWSClient
func NewHTNWSClient() *NWSClient {
	var htnwsc *NWSClient
	c := nconf.GetConfig()
	scheme := c.GetString(NameHTWSC + ".nwsc.scheme")
	address := c.GetString(NameHTWSC + ".nwsc.host")
	path := c.GetString(NameHTWSC + ".nwsc.path")
	log.Printf("################ HTNWSClient[%s] start...", NameHTWSC)
	htnwsc, _ = NewInstanceWSC(NameHTWSC, scheme, address, path)
	// htnwsc, _ = NewInstanceWSC(NameHTWSC, "wss", "stream.binance.com:9443", "/ws/btcusdt@trade")
	return htnwsc
}

// StartHTNWSClient start
func (nwsc *NWSClient) StartHTNWSClient() {
	// Thread receive message.
	go nwsc.recvHT()
	// Thread send message.
	//go nwsc.sendHT()
}
