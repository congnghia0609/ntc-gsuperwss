/**
 *
 * @author nghiatc
 * @since Aug 8, 2018
 */

package nwsc

import (
	"fmt"
	"log"
	"ntc-gsuperwss/nwss"
	"ntc-gsuperwss/util"
	"time"

	"github.com/congnghia0609/ntc-gconf/nconf"
	"github.com/gorilla/websocket"
)

func (nwsc *NWSClient) recvCR() {
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
				// log.Printf("recvCR: %s", message)
				if len(message) > 0 {
					// CRNWSServer
					crnwss := nwss.GetInstanceCR(nwss.NameCRNWSS)
					if crnwss != nil {
						crnwss.BroadcastMsgByte(message)
					}
				}
			}
		},
		Catch: func(e util.Exception) {
			log.Printf("nwsc.recvCR Caught %v\n", e)
		},
		Finally: func() {
			//log.Println("Finally...")
		},
	}.Do()
}

func (nwsc *NWSClient) sendCR() {
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
			log.Printf("nwsc.sendCR Caught %v\n", e)
		},
		Finally: func() {
			//log.Println("Finally...")
		},
	}.Do()
}

// NewCRNWSClient new instance of NWSClient
func NewCRNWSClient() *NWSClient {
	var crnwsc *NWSClient
	c := nconf.GetConfig()
	scheme := c.GetString(NameCRWSC + ".nwsc.scheme")
	address := c.GetString(NameCRWSC + ".nwsc.host")
	path := c.GetString(NameCRWSC + ".nwsc.path")
	log.Printf("################ CRNWSClient[%s] start...", NameCRWSC)
	crnwsc, _ = NewInstanceWSC(NameCRWSC, scheme, address, path)
	// crnwsc, _ = NewInstanceWSC(NameCRWSC, "ws", address, "/dataws/depth")
	// crnwsc, _ = NewInstanceWSC(NameCRWSC, "ws", "localhost:15501", "/ws/v1/dp/ETH_BTC")
	// crnwsc, _ = NewInstanceWSC(NameCRWSC, "wss", "engine2.kryptono.exchange", "/ws/v1/dp/ETH_BTC")
	return crnwsc
}

// StartCRNWSClient start
func (nwsc *NWSClient) StartCRNWSClient() {
	// Thread receive message.
	go nwsc.recvCR()
	// Thread send message.
	//go nwsc.sendCR()
}
