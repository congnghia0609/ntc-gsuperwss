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
				log.Printf("recvDP: %s", message)
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
					//data := `{"a":[],"b":[["379.11400000", "0.03203000"]],"s":"ETH_BTC","t":"` + fmt.Sprint(msec) + `","e":"depthUpdate"}`
					// `{"e":"depthUpdate","E":1636174562247,"s":"BTCUSDT","U":14824495332,"u":14824495377,"b":[["61217.72000000","0.84279000"],["61197.54000000","0.00000000"],["61197.53000000","0.00000000"],["61167.77000000","0.02279000"],["61112.00000000","0.00000000"],["60800.00000000","8.41281000"],["59510.00000000","0.40314000"]],"a":[["61217.73000000","0.60274000"],["61223.32000000","0.00000000"],["61224.95000000","0.00000000"],["61224.96000000","0.00000000"],["61225.43000000","0.03000000"],["61225.69000000","0.08000000"],["61225.87000000","0.00000000"],["61227.16000000","0.00000000"],["61227.90000000","0.25733000"],["61228.47000000","0.00000000"],["61228.81000000","0.50000000"],["61229.23000000","0.00000000"],["61229.93000000","0.00000000"],["61231.22000000","0.00000000"],["61234.26000000","0.00000000"],["61234.27000000","0.00327000"],["61246.91000000","0.00000000"],["61313.72000000","0.99711000"],["61315.07000000","0.00000000"],["61363.82000000","0.00000000"],["61368.72000000","0.09564000"],["61407.05000000","0.00000000"],["61913.52000000","2.51298000"]]}`
					data := `{"e":"depthUpdate","E":` + fmt.Sprint(msec) + `,"s":"BTCUSDT","U":14824495332,"u":14824495377,"b":[["61217.72000000","0.84279000"],["61197.54000000","0.00000000"],["61197.53000000","0.00000000"],["61167.77000000","0.02279000"],["61112.00000000","0.00000000"],["60800.00000000","8.41281000"],["59510.00000000","0.40314000"]],"a":[["61217.73000000","0.60274000"],["61223.32000000","0.00000000"],["61224.95000000","0.00000000"],["61224.96000000","0.00000000"],["61225.43000000","0.03000000"],["61225.69000000","0.08000000"],["61225.87000000","0.00000000"],["61227.16000000","0.00000000"],["61227.90000000","0.25733000"],["61228.47000000","0.00000000"],["61228.81000000","0.50000000"],["61229.23000000","0.00000000"],["61229.93000000","0.00000000"],["61231.22000000","0.00000000"],["61234.26000000","0.00000000"],["61234.27000000","0.00327000"],["61246.91000000","0.00000000"],["61313.72000000","0.99711000"],["61315.07000000","0.00000000"],["61363.82000000","0.00000000"],["61368.72000000","0.09564000"],["61407.05000000","0.00000000"],["61913.52000000","2.51298000"]]}`
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
	// dpwsc, _ = NewInstanceWSC(NameDPWSC, "wss", "stream.binance.com:9443", "/ws/btcusdt@depth")
	// dpwsc, _ = NewInstanceWSC(NameDPWSC, "ws", "127.0.0.1:15501", "/ws/v1/dp/btcusdt")
	return dpwsc
}

// StartDPNWSClient start
func (nwsc *NWSClient) StartDPNWSClient() {
	// Thread receive message.
	go nwsc.recvDP()
	// Thread send message.
	//go nwsc.sendDP()
}
