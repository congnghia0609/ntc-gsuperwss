/**
 *
 * @author nghiatc
 * @since Sep 30, 2020
 */

package nwss

import (
	"github.com/congnghia0609/ntc-gsuperwss/util"
	"log"
	"net"
	"net/http"

	"github.com/congnghia0609/ntc-gconf/nconf"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

// TKNWSServer struct
type TKNWSServer struct {
	name    string
	epoller *NEpoll
	// Inbound message from the clients.
	broadcast chan []byte
}

// mapInstanceTK management instance
var mapInstanceTK = make(map[string]*TKNWSServer)

// GetInstanceTK get instance TK
func GetInstanceTK(name string) *TKNWSServer {
	return mapInstanceTK[name]
}

// GetName get name
func (nwss *TKNWSServer) GetName() string {
	return nwss.name
}

// GetNEpoll get NEpoll
func (nwss *TKNWSServer) GetNEpoll() *NEpoll {
	return nwss.epoller
}

// NewTKNWSServer new TKNWSServer
func NewTKNWSServer(name string) *TKNWSServer {
	// New NEpoll
	epoller, err := MkNEpoll()
	if err != nil {
		panic(err)
	}
	instance := &TKNWSServer{name: name, epoller: epoller, broadcast: make(chan []byte)}
	mapInstanceTK[name] = instance
	return instance
}

// BroadcastMsgByte broadcast msg byte
func (nwss *TKNWSServer) BroadcastMsgByte(message []byte) {
	util.TCF{
		Try: func() {
			if len(message) > 0 {
				// log.Printf("message: %s", message)
				nwss.broadcast <- message
			}
		},
		Catch: func(e util.Exception) {
			log.Printf("TKNWSServer.BroadcastMsgByte Caught: %v\n", e)
		},
		Finally: func() {
			//log.Println("Finally...")
		},
	}.Do()
}

// closeConn close connection
func (nwss *TKNWSServer) closeConn(conn net.Conn) {
	if err := nwss.epoller.Remove(conn); err != nil {
		log.Printf("Failed to remove: %v", err)
	}
	conn.Close()
}

// readClientData read client data
func (nwss *TKNWSServer) readClientData() {
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
func (nwss *TKNWSServer) broadcastData() {
	for {
		select {
		case message := <-nwss.broadcast:
			util.TCF{
				Try: func() {
					if len(message) > 0 {
						// log.Printf("message: %s", message)
						for _, conn := range nwss.epoller.connections {
							if conn != nil {
								err := wsutil.WriteServerMessage(conn, ws.OpText, message)
								if err != nil {
									log.Printf("Send to client failed: %v", err)
								}
							}
						}
					}
				},
				Catch: func(e util.Exception) {
					log.Printf("TKNWSServer.broadcast Caught: %v\n", e)
				},
				Finally: func() {
					//log.Println("Finally...")
				},
			}.Do()
		}
	}
}

// wsTKHandler ws handler
func (nwss *TKNWSServer) wsTKHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade connection
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		return
	}
	// Add connection to NEpoll
	if _, err := nwss.epoller.Add(conn); err != nil {
		log.Printf("Failed to add connection: %v", err)
		conn.Close()
		return
	}
	// Push message connected successfully to client.
	msgsc := `{"err":0,"msg":"Connected successfully"}`
	errw := wsutil.WriteServerMessage(conn, ws.OpText, []byte(msgsc))
	if errw != nil {
		log.Printf("Send to client failed: %v", errw)
	}
}

// Start websocket server
func (nwss *TKNWSServer) Start() {
	// read config
	c := nconf.GetConfig()
	host := c.GetString(nwss.name + ".nwss.host")
	// path := c.GetString(nwss.name + ".nwss.path")

	go nwss.broadcastData()
	go nwss.readClientData()

	http.HandleFunc("/ws/v1/tk", nwss.wsTKHandler)
	log.Printf("======= TKNWSServer[%s] is running on host: %s\n", nwss.name, host)
	if err := http.ListenAndServe(host, nil); err != nil {
		log.Fatal(err)
	}
}
