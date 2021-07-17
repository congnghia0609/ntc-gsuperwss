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

func (nwsc *NWSClient) recvCS() {
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
				// log.Printf("recvCS: %s", message)
				if len(message) > 0 {
					// CSNWSServer
					csnwss := nwss.GetInstanceCS(nwss.NameCSNWSS)
					if csnwss != nil {
						csnwss.BroadcastMsgByte(message)
					}
				}
			}
		},
		Catch: func(e util.Exception) {
			log.Printf("nwsc.recvCS Caught %v\n", e)
		},
		Finally: func() {
			//log.Println("Finally...")
		},
	}.Do()
}

func (nwsc *NWSClient) sendCS() {
	util.TCF{
		Try: func() {
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()

			for {
				select {
				case t := <-ticker.C:
					//err := nws.conn.WriteMessage(websocket.TextMessage, []byte(t.String()))
					msec := t.UnixNano() / 1000000
					///// 1. Candlesticks Data.
					data := `{"tt":"1h","s":"ETH_BTC","t":` + fmt.Sprint(msec) + `,"e":"kline","k":{"c":"0.00028022","t":1533715200000,"v":"905062.00000000","h":"0.00028252","l":"0.00027787","o":"0.00027919"}}`
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
			log.Printf("nwsc.sendCS Caught %v\n", e)
		},
		Finally: func() {
			//log.Println("Finally...")
		},
	}.Do()
}

// NewCSNWSClient new instance of NWSClient
func NewCSNWSClient() *NWSClient {
	var csnwsc *NWSClient
	c := nconf.GetConfig()
	scheme := c.GetString(NameCSWSC + ".nwsc.scheme")
	address := c.GetString(NameCSWSC + ".nwsc.host")
	path := c.GetString(NameCSWSC + ".nwsc.path")
	log.Printf("################ CSNWSClient[%s] start...", NameCSWSC)
	csnwsc, _ = NewInstanceWSC(NameCSWSC, scheme, address, path)
	// csnwsc, _ = NewInstanceWSC(NameCSWSC, "ws", address, "/dataws/depth")
	// csnwsc, _ = NewInstanceWSC(NameCSWSC, "ws", "localhost:15501", "/ws/v1/dp/ETH_BTC")
	// csnwsc, _ = NewInstanceWSC(NameCSWSC, "wss", "engine2.kryptono.exchange", "/ws/v1/dp/ETH_BTC")
	return csnwsc
}

// StartCSNWSClient start
func (nwsc *NWSClient) StartCSNWSClient() {
	// Thread receive message.
	go nwsc.recvCS()
	// Thread send message.
	//go nwsc.sendCS()
}
