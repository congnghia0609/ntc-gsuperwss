/**
 *
 * @author nghiatc
 * @since Sep 30, 2020
 */

package nwss

import (
	"log"
	"net/http"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

var epoller *epoll

func wsHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade connection
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		return
	}
	if err := epoller.Add(conn); err != nil {
		log.Printf("Failed to add connection: %v", err)
		conn.Close()
	}
}

// Start websocket server
func Start() {
	for {
		connections, err := epoller.Wait()
		if err != nil {
			log.Printf("Failed to epoll wait: %v", err)
			continue
		}
		for _, conn := range connections {
			if conn == nil {
				break
			}
			// msg, op, err := wsutil.ReadClientData(conn)
			_, _, err := wsutil.ReadClientData(conn)
			if err != nil {
				if err := epoller.Remove(conn); err != nil {
					log.Printf("Failed to remove: %v", err)
				}
				conn.Close()
			} else {
				// This is commented out since in demo usage,
				// stdout is showing messages sent from > 1M connections at very high rate
				// log.Printf("msg: %s", string(msg))
				// err := wsutil.WriteServerMessage(conn, op, msg)
				// if err != nil {
				// 	log.Printf("Send to client failed: %v", err)
				// }
			}
		}
	}
}

// Run demo
func Run() {
	// Start epoll
	var err error
	epoller, err = MkEpoll()
	if err != nil {
		panic(err)
	}

	go Start()

	http.HandleFunc("/", wsHandler)
	log.Printf("======= NWSServer is running on host: 0.0.0.0:8000")
	if err := http.ListenAndServe("0.0.0.0:8000", nil); err != nil {
		log.Fatal(err)
	}
}
