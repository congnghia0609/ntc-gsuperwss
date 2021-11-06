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

// CSNWSServer struct
type CSNWSServer struct {
	name    string
	epoller *NEpoll
	// Inbound message from the clients.
	broadcast chan []byte
	// Map Symbol_TypeTime Conn
	symbolTTConn map[string]map[int]bool
}

// mapInstanceCS management instance
var mapInstanceCS = make(map[string]*CSNWSServer)

// GetInstanceCS get instance CS
func GetInstanceCS(name string) *CSNWSServer {
	return mapInstanceCS[name]
}

// GetName get name
func (nwss *CSNWSServer) GetName() string {
	return nwss.name
}

// GetNEpoll get NEpoll
func (nwss *CSNWSServer) GetNEpoll() *NEpoll {
	return nwss.epoller
}

// NewCSNWSServer new CSNWSServer
func NewCSNWSServer(name string) *CSNWSServer {
	// New NEpoll
	epoller, err := MkNEpoll()
	if err != nil {
		panic(err)
	}
	instance := &CSNWSServer{
		name:         name,
		epoller:      epoller,
		broadcast:    make(chan []byte),
		symbolTTConn: make(map[string]map[int]bool),
	}
	mapInstanceCS[name] = instance
	return instance
}

// BroadcastMsgByte broadcast msg byte
func (nwss *CSNWSServer) BroadcastMsgByte(message []byte) {
	util.TCF{
		Try: func() {
			if len(message) > 0 {
				// log.Printf("message: %s", message)
				nwss.broadcast <- message
			}
		},
		Catch: func(e util.Exception) {
			log.Printf("CSNWSServer.BroadcastMsgByte Caught: %v\n", e)
		},
		Finally: func() {
			//log.Println("Finally...")
		},
	}.Do()
}

// closeConn close connection
func (nwss *CSNWSServer) closeConn(conn net.Conn) {
	if err := nwss.epoller.Remove(conn); err != nil {
		log.Printf("Failed to remove: %v", err)
	}
	conn.Close()
}

// readClientData read client data
func (nwss *CSNWSServer) readClientData() {
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
			//msg, _, err := wsutil.ReadClientData(conn)
			_, _, err := wsutil.ReadClientData(conn)
			if err != nil {
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
				//nwss.broadcast <- msg

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
func (nwss *CSNWSServer) broadcastData() {
	for {
		select {
		case message := <-nwss.broadcast:
			util.TCF{
				Try: func() {
					if len(message) > 0 {
						// log.Printf("message: %s", message)
						var data map[string]interface{}
						json.Unmarshal([]byte(message), &data)
						if data["s"] != nil && data["k"] != nil {
							symbol := data["s"].(string)
							k := data["k"].(map[string]interface{})
							if k["i"] != nil {
								tt := k["i"].(string)
								// log.Printf("HubLevel2.broadcast {symbol=%s,typeTime=%s}", symbol, tt)
								if len(symbol) > 0 && len(tt) > 0 {
									symbol = strings.ToLower(symbol)
									tt = strings.ToLower(tt)
									key := symbol + "@" + tt
									for fd := range nwss.symbolTTConn[key] {
										conn := nwss.epoller.GetConn(fd)
										if conn != nil {
											// Send message to client
											err := wsutil.WriteServerMessage(conn, ws.OpText, message)
											if err != nil {
												log.Printf("Send to client failed: %v", err)
											}
										} else {
											// Delete fd from map symbolTTConn
											delete(nwss.symbolTTConn[key], fd)
										}
									}
								}
							}
						}
					}
				},
				Catch: func(e util.Exception) {
					log.Printf("CSNWSServer.broadcast Caught: %v\n", e)
				},
				Finally: func() {
					//log.Println("Finally...")
				},
			}.Do()
		}
	}
}

// wsCSHandler ws handler
func (nwss *CSNWSServer) wsCSHandler(w http.ResponseWriter, r *http.Request) {
	// Validate
	pathURI := r.RequestURI
	log.Printf("=======pathURI: %s", pathURI)
	vars := mux.Vars(r)
	stt := vars["STT"]
	if len(stt) <= 0 {
		return
	}
	stt = strings.ToLower(stt)
	arrSTT := strings.Split(stt, "@")
	if len(arrSTT) != 2 {
		return
	}
	symbol := strings.ToLower(arrSTT[0])
	tt := strings.ToLower(arrSTT[1])
	log.Printf("=======symbol: %s, typeTime: %s", symbol, tt)
	_, ok1 := MapSymbol[symbol]
	_, ok2 := TypeTime[tt]
	if !(ok1 && ok2) {
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
	// Add connection to map symbolTTConn
	if _, ok := nwss.symbolTTConn[stt]; ok {
		nwss.symbolTTConn[stt][fd] = true
	} else {
		nwss.symbolTTConn[stt] = make(map[int]bool)
		nwss.symbolTTConn[stt][fd] = true
	}
	// Push message connected successfully to client.
	msgsc := `{"err":0,"msg":"Connected successfully"}`
	errw := wsutil.WriteServerMessage(conn, ws.OpText, []byte(msgsc))
	if errw != nil {
		log.Printf("Send to client failed: %v", errw)
	}
}

// Start websocket server
func (nwss *CSNWSServer) Start() {
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
	rt.HandleFunc("/ws/v1/cs/{STT}", nwss.wsCSHandler)
	httpsm.Handle("/", rt)

	log.Printf("======= CSNWSServer[%s] is running on host: %s\n", nwss.name, host)
	if err := http.ListenAndServe(host, httpsm); err != nil {
		log.Fatal(err)
	}
}
