/**
 *
 * @author nghiatc
 * @since Sep 30, 2020
 */

package nwss

import (
	"encoding/json"
	"github.com/congnghia0609/ntc-gsuperwss/util"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/congnghia0609/ntc-gconf/nconf"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/gorilla/mux"
)

// CRNWSServer struct
type CRNWSServer struct {
	name    string
	epoller *NEpoll
	// Inbound message from the clients.
	broadcast chan []byte
	// Map Symbol Conn
	symbolConn map[string]map[int]bool
}

// mapInstanceCR management instance
var mapInstanceCR = make(map[string]*CRNWSServer)

// GetInstanceCR get instance CR
func GetInstanceCR(name string) *CRNWSServer {
	return mapInstanceCR[name]
}

// GetName get name
func (nwss *CRNWSServer) GetName() string {
	return nwss.name
}

// GetNEpoll get NEpoll
func (nwss *CRNWSServer) GetNEpoll() *NEpoll {
	return nwss.epoller
}

// NewCRNWSServer new CRNWSServer
func NewCRNWSServer(name string) *CRNWSServer {
	// New NEpoll
	epoller, err := MkNEpoll()
	if err != nil {
		panic(err)
	}
	instance := &CRNWSServer{
		name:       name,
		epoller:    epoller,
		broadcast:  make(chan []byte),
		symbolConn: make(map[string]map[int]bool),
	}
	mapInstanceCR[name] = instance
	return instance
}

// BroadcastMsgByte broadcast msg byte
func (nwss *CRNWSServer) BroadcastMsgByte(message []byte) {
	util.TCF{
		Try: func() {
			if len(message) > 0 {
				// log.Printf("message: %s", message)
				nwss.broadcast <- message
			}
		},
		Catch: func(e util.Exception) {
			log.Printf("CRNWSServer.BroadcastMsgByte Caught: %v\n", e)
		},
		Finally: func() {
			//log.Println("Finally...")
		},
	}.Do()
}

// closeConn close connection
func (nwss *CRNWSServer) closeConn(conn net.Conn) {
	if err := nwss.epoller.Remove(conn); err != nil {
		log.Printf("Failed to remove: %v", err)
	}
	conn.Close()
}

// readClientData read client data
func (nwss *CRNWSServer) readClientData() {
	for {
		connections, err := nwss.epoller.Wait()
		if err != nil {
			// log.Printf("Failed to epoll wait: %v", err)
			continue
		}
		for _, conn := range connections {
			if conn == nil {
				continue
			}
			// msg, op, err := wsutil.ReadClientData(conn)
			// msg, _, err := wsutil.ReadClientData(conn)
			_, _, err := wsutil.ReadClientData(conn)
			if err != nil {
				// if err := nwss.epoller.Remove(conn); err != nil {
				// 	log.Printf("Failed to remove: %v", err)
				// }
				// conn.Close()
				nwss.closeConn(conn)
			} else {
				/// Process Business Here.

				// Not receive message from Client. {"msg":"Message invalid","err":-1}
				msg := `{"err":-1,"msg":"Message invalid"}`
				err := wsutil.WriteServerMessage(conn, ws.OpText, []byte(msg))
				if err != nil {
					log.Printf("Send to client failed: %v", err)
				}

				// // broadcast message
				// nwss.broadcast <- msg

				// This is commented out since in demo usage,
				// stdout is showing messages sent from > 1M connections at very high rate
				// log.Printf("msg: %s", string(msg))
				// log.Printf("msg: %s | op: %v | err: %v", string(msg), op, err)
				// err := wsutil.WriteServerMessage(conn, op, msg)
				// if err != nil {
				// 	log.Printf("Send to client failed: %v", err)
				// }
			}
		}
	}
}

// broadcastData broadcast data to client
func (nwss *CRNWSServer) broadcastData() {
	for {
		select {
		case message := <-nwss.broadcast:
			util.TCF{
				Try: func() {
					if len(message) > 0 {
						// log.Printf("message: %s", message)
						var data map[string]interface{}
						json.Unmarshal([]byte(message), &data)
						if data["et"] != nil {
							et := data["et"].(string)
							// broadcast Ticker24h
							if strings.EqualFold("tk", et) {
								for _, conn := range nwss.epoller.connections {
									if conn != nil {
										err := wsutil.WriteServerMessage(conn, ws.OpText, message)
										if err != nil {
											log.Printf("Send to client failed: %v", err)
										}
									}
								}
							} else if strings.EqualFold("dp", et) || strings.EqualFold("ht", et) {
								// broadcast DepthPrice & HistoryTrade
								if data["s"] != nil {
									symbol := data["s"].(string)
									// log.Printf("CRNWSServer.broadcast.symbol=%s", symbol)
									if len(symbol) > 0 {
										symbol = strings.ToLower(symbol)
										for fd := range nwss.symbolConn[symbol] {
											conn := nwss.epoller.GetConn(fd)
											if conn != nil {
												// Send message to client
												err := wsutil.WriteServerMessage(conn, ws.OpText, message)
												if err != nil {
													log.Printf("Send to client failed: %v", err)
												}
											} else {
												// Delete fd from map symbolConn
												delete(nwss.symbolConn[symbol], fd)
											}
										}
									}
								}
							}
						}
					}
				},
				Catch: func(e util.Exception) {
					log.Printf("CRNWSServer.broadcast Caught: %v\n", e)
				},
				Finally: func() {
					//log.Println("Finally...")
				},
			}.Do()
		}
	}
}

// wsCRHandler ws handler
func (nwss *CRNWSServer) wsCRHandler(w http.ResponseWriter, r *http.Request) {
	// Validate
	pathURI := r.RequestURI
	log.Printf("=======pathURI: %s", pathURI)
	vars := mux.Vars(r)
	symbol := vars["symbol"]
	if len(symbol) <= 0 {
		return
	}
	symbol = strings.ToLower(symbol)
	log.Printf("=======symbol: %s", symbol)
	if _, ok := MapSymbol[symbol]; !ok {
		return
	}

	// Upgrade connection
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		return
	}
	// Add connection to NEpoll
	fd, err := nwss.epoller.Add(conn)
	if err != nil {
		log.Printf("Failed to add connection: %v", err)
		conn.Close()
		return
	}
	// Add connection to map symbolConn
	if _, ok := nwss.symbolConn[symbol]; ok {
		nwss.symbolConn[symbol][fd] = true
	} else {
		nwss.symbolConn[symbol] = make(map[int]bool)
		nwss.symbolConn[symbol][fd] = true
	}
	// Push message connected successfully to client.
	msgsc := `{"err":0,"msg":"Connected successfully"}`
	errw := wsutil.WriteServerMessage(conn, ws.OpText, []byte(msgsc))
	if errw != nil {
		log.Printf("Send to client failed: %v", errw)
	}
}

// Start websocket server
func (nwss *CRNWSServer) Start() {
	// read config
	c := nconf.GetConfig()
	host := c.GetString(nwss.name + ".nwss.host")
	// path := c.GetString(nwss.name + ".nwss.path")

	go nwss.broadcastData()
	go nwss.readClientData()

	// NewServeMux
	httpsm := http.NewServeMux()
	// Setup Handlers.
	rt := mux.NewRouter()
	rt.HandleFunc("/ws/v1/cr/{symbol}", nwss.wsCRHandler)
	httpsm.Handle("/", rt)

	log.Printf("======= CRNWSServer[%s] is running on host: %s\n", nwss.name, host)
	if err := http.ListenAndServe(host, httpsm); err != nil {
		log.Fatal(err)
	}
}
