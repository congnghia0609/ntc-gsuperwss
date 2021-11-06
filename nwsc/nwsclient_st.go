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

func (nwsc *NWSClient) recvST() {
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
				log.Printf("recvST: %s", message)
				if len(message) > 0 {
					stwss := nwss.GetInstanceST(nwss.NameSTNWSS)
					if stwss != nil {
						stwss.BroadcastMsgByte(message)
					}
				}
			}
		},
		Catch: func(e util.Exception) {
			log.Printf("nwsc.recvST Caught %v\n", e)
		},
		Finally: func() {
			//log.Println("Finally...")
		},
	}.Do()
}

func (nwsc *NWSClient) sendST() {
	util.TCF{
		Try: func() {
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()
			for {
				select {
				case t := <-ticker.C:
					//err := nws.conn.WriteMessage(websocket.TextMessage, []byte(t.String()))
					msec := t.UnixNano() / 1000000
					///// 1. Stream Data.
					//`{"stream":"btcusdt@aggTrade","data":{"e":"aggTrade","E":1636177718119,"s":"BTCUSDT","a":983208851,"p":"61384.88000000","q":"0.00244000","f":1131824289,"l":1131824289,"T":1636177718118,"m":true,"M":true}}`
					data := `{"stream":"btcusdt@aggTrade","data":{"e":"aggTrade","E":` + fmt.Sprint(msec) + `,"s":"BTCUSDT","a":983208851,"p":"61384.88000000","q":"0.00244000","f":1131824289,"l":1131824289,"T":1636177718118,"m":true,"M":true}}`
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
			log.Printf("nwsc.sendST Caught %v\n", e)
		},
		Finally: func() {
			//log.Println("Finally...")
		},
	}.Do()
}

// NewSTNWSClient new instance STWSClient of NWSClient
func NewSTNWSClient() *NWSClient {
	var stwsc *NWSClient
	c := nconf.GetConfig()
	scheme := c.GetString(NameSTWSC + ".nwsc.scheme")
	address := c.GetString(NameSTWSC + ".nwsc.host")
	path := c.GetString(NameSTWSC + ".nwsc.path")
	log.Printf("################ STNWSClient[%s] start...", NameSTWSC)
	stwsc, _ = NewInstanceWSC(NameSTWSC, scheme, address, path)
	// stwsc, _ = NewInstanceWSC(NameSTWSC, "wss", "stream.binance.com", "/stream")
	// stwsc, _ = NewInstanceWSC(NameSTWSC, "ws", "127.0.0.1:19001", "/stream")
	return stwsc
}

// StartSTNWSClient start
func (nwsc *NWSClient) StartSTNWSClient() {
	// Thread receive message.
	go nwsc.recvST()
	// Thread send message.
	//go nwsc.sendST()
	// Send message subscribe.
	msg := "{\"method\":\"SUBSCRIBE\",\"params\":[\"!miniTicker@arr@3000ms\",\"btcusdt@aggTrade\",\"btcusdt@depth\",\"btcusdt@kline_1d\"],\"id\":1}"
	nwsc.conn.WriteMessage(websocket.TextMessage, []byte(msg))
}
