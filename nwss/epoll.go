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

type epoll struct {
	fd          int
	connections map[int]net.Conn
	lock        *sync.RWMutex
}

// MkEpoll new epoll.
// Returns *epoll and any write error encountered.
func MkEpoll() (*epoll, error) {
	fd, err := unix.EpollCreate1(0)
	if err != nil {
		return nil, err
	}
	return &epoll{
		fd:          fd,
		lock:        &sync.RWMutex{},
		connections: make(map[int]net.Conn),
	}, nil
}

// websocketFD get ID from net.Conn
func websocketFD(conn net.Conn) int {
	tcpConn := reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn")
	fdVal := tcpConn.FieldByName("fd")
	pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")
	return int(pfdVal.FieldByName("Sysfd").Int())
}

// Add add connection to epoll
func (e *epoll) Add(conn net.Conn) (int, error) {
	// Extract file descriptor associated with the connection
	fd := websocketFD(conn)
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

// Get get conn from epoll
func (e *epoll) Get(fd int) net.Conn {
	return e.connections[fd]
}

// Remove delete conn from epoll
func (e *epoll) Remove(conn net.Conn) error {
	fd := websocketFD(conn)
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
func (e *epoll) Wait() ([]net.Conn, error) {
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
