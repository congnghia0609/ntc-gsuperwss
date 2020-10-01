/**
 *
 * @author nghiatc
 * @since Sep 30, 2020
 */

package nwss

import (
	"log"
	"net"
	"reflect"
	"sync"
	"syscall"

	"golang.org/x/sys/unix"
)

// NEpoll struct
type NEpoll struct {
	fd          int
	connections map[int]net.Conn
	lock        *sync.RWMutex
}

// MkNEpoll new NEpoll.
// Returns *NEpoll and any write error encountered.
func MkNEpoll() (*NEpoll, error) {
	fd, err := unix.EpollCreate1(0)
	if err != nil {
		return nil, err
	}
	return &NEpoll{
		fd:          fd,
		lock:        &sync.RWMutex{},
		connections: make(map[int]net.Conn),
	}, nil
}

// FDWebsocket get ID from net.Conn
func FDWebsocket(conn net.Conn) int {
	tcpConn := reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn")
	fdVal := tcpConn.FieldByName("fd")
	pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")
	return int(pfdVal.FieldByName("Sysfd").Int())
}

// Add add connection to epoll
func (e *NEpoll) Add(conn net.Conn) (int, error) {
	// Extract file descriptor associated with the connection
	fd := FDWebsocket(conn)
	err := unix.EpollCtl(e.fd, syscall.EPOLL_CTL_ADD, fd, &unix.EpollEvent{Events: unix.POLLIN | unix.POLLHUP, Fd: int32(fd)})
	if err != nil {
		return fd, err
	}
	e.lock.Lock()
	defer e.lock.Unlock()
	e.connections[fd] = conn
	if len(e.connections)%100 == 0 {
		log.Printf("Total number of connections: %v", len(e.connections))
	}
	return fd, nil
}

// GetConn get conn from NEpoll
func (e *NEpoll) GetConn(fd int) net.Conn {
	return e.connections[fd]
}

// GetTotalConn get total connections
func (e *NEpoll) GetTotalConn() int {
	return len(e.connections)
}

// Remove delete conn from NEpoll
func (e *NEpoll) Remove(conn net.Conn) error {
	fd := FDWebsocket(conn)
	err := unix.EpollCtl(e.fd, syscall.EPOLL_CTL_DEL, fd, nil)
	if err != nil {
		return err
	}
	e.lock.Lock()
	defer e.lock.Unlock()
	delete(e.connections, fd)
	if len(e.connections)%100 == 0 {
		log.Printf("Total number of connections: %v", len(e.connections))
	}
	return nil
}

// Wait process batch 100 events from unix.EpollWait
func (e *NEpoll) Wait() ([]net.Conn, error) {
	events := make([]unix.EpollEvent, 100)
	n, err := unix.EpollWait(e.fd, events, 100)
	if err != nil {
		return nil, err
	}
	e.lock.Lock()
	defer e.lock.Unlock()
	var connections []net.Conn
	for i := 0; i < n; i++ {
		conn := e.connections[int(events[i].Fd)]
		connections = append(connections, conn)
	}
	return connections, nil
}
