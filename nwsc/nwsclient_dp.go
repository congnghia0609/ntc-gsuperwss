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

func (nwsc *NWSClient) recvDP() {
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
				// log.Printf("recvDP: %s", message)
				if len(message) > 0 {
					// DPWSServer
					dpws := nwss.GetInstanceDP(nwss.NameDPNWSS)
					if dpws != nil {
						dpws.BroadcastMsgByte(message)
					}
				}
			}
		},
		Catch: func(e util.Exception) {
			log.Printf("nwsc.recvDP Caught %v\n", e)
		},
		Finally: func() {
			//log.Println("Finally...")
		},
	}.Do()
}

func (nwsc *NWSClient) sendDP() {
	util.TCF{
		Try: func() {
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()

			for {
				select {
				case t := <-ticker.C:
					//err := nws.conn.WriteMessage(websocket.TextMessage, []byte(t.String()))
					msec := t.UnixNano() / 1000000
					///// 1. DepthPrice Data.
					data := `{"a":[],"b":[["379.11400000", "0.03203000"]],"s":"ETH_BTC","t":"` + fmt.Sprint(msec) + `","e":"depthUpdate"}`
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
			log.Printf("nwsc.sendDP Caught %v\n", e)
		},
		Finally: func() {
			//log.Println("Finally...")
		},
	}.Do()
}

// NewDPNWSClient new instance of NWSClient
func NewDPNWSClient() *NWSClient {
	var dpwsc *NWSClient
	c := nconf.GetConfig()
	scheme := c.GetString(NameDPWSC + ".nwsc.scheme")
	address := c.GetString(NameDPWSC + ".nwsc.host")
	path := c.GetString(NameDPWSC + ".nwsc.path")
	log.Printf("################ DPNWSClient[%s] start...", NameDPWSC)
	dpwsc, _ = NewInstanceWSC(NameDPWSC, scheme, address, path)
	// dpwsc, _ = NewInstanceWSC(NameDPWSC, "ws", address, "/dataws/depth")
	// dpwsc, _ = NewInstanceWSC(NameDPWSC, "ws", "localhost:15501", "/ws/v1/dp/ETH_BTC")
	// dpwsc, _ = NewInstanceWSC(NameDPWSC, "wss", "engine2.kryptono.exchange", "/ws/v1/dp/ETH_BTC")
	return dpwsc
}

// StartDPNWSClient start
func (nwsc *NWSClient) StartDPNWSClient() {
	// Thread receive message.
	go nwsc.recvDP()
	// Thread send message.
	//go nwsc.sendDP()
}
