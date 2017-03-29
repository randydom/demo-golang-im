package tcp

import (
	"net"
	"sync/atomic"
	"time"
	"logger"
)

const (
	kServerConf_SendBufferSize = 1024
	kServerConn                = 0
	kClientConn                = 1
)

type TCPNetworkConf struct {
	SendBufferSize int
}

type TCPNetwork struct {
	streamProtocol  IStreamProtocol
	eventQueue      chan *ConnEvent
	listener        net.Listener
	Conf            TCPNetworkConf
	connIdForServer int
	connIdForClient int
	connsForServer  map[int]*Connection
	connsForClient  map[int]*Connection
	shutdownFlag    int32
	readTimeoutSec  int
}

func NewTCPNetwork(eventQueueSize int, sp IStreamProtocol, buffSize int) *TCPNetwork {
	s := &TCPNetwork{}
	s.eventQueue = make(chan *ConnEvent, eventQueueSize)
	s.streamProtocol = sp
	s.connsForServer = make(map[int]*Connection)
	s.connsForClient = make(map[int]*Connection)
	s.shutdownFlag = 0
	s.Conf.SendBufferSize = buffSize
	return s
}

// Push implements the IEventQueue interface
func (t *TCPNetwork) Push(evt *ConnEvent) {
	if nil == t.eventQueue {
		return
	}

	//	push timeout
	select {
	case t.eventQueue <- evt:
		{

		}
	case <-time.After(time.Second * 5):
		{
			evt.Conn.Close()
		}
	}

}

// Pop the event in event queue
func (t *TCPNetwork) Pop() *ConnEvent {
	evt, ok := <-t.eventQueue
	if !ok {
		//	event queue already closed
		return nil
	}

	return evt
}

// GetEventQueue get the event queue channel
func (t *TCPNetwork) GetEventQueue() <-chan *ConnEvent {
	return t.eventQueue
}

// Listen an address to accept client connection
func (t *TCPNetwork) Listen(addr string) error {
	ls, err := net.Listen("tcp", addr)
	if nil != err {
		return err
	}

	//	accept
	t.listener = ls
	go t.acceptRoutine()
	return nil
}

// Connect the remote server
func (t *TCPNetwork) Connect(addr string) (*Connection, error) {
	tcpConn, err := net.Dial("tcp", addr)
	if nil != err {
		return nil, err
	}

	connection := t.createConn(tcpConn)
	connection.From = kClientConn
	connection.Run()
	connection.Init()

	return connection, nil
}

func (t *TCPNetwork) GetStreamProtocol() IStreamProtocol {
	return t.streamProtocol
}

func (t *TCPNetwork) SetStreamProtocol(sp IStreamProtocol) {
	t.streamProtocol = sp
}

func (t *TCPNetwork) GetReadTimeoutSec() int {
	return t.readTimeoutSec
}

func (t *TCPNetwork) SetReadTimeoutSec(sec int) {
	t.readTimeoutSec = sec
}

func (t *TCPNetwork) DisconnectAllConnectionsServer() {
	for k, c := range t.connsForServer {
		c.Close()
		delete(t.connsForServer, k)
	}
}

func (t *TCPNetwork) DisconnectAllConnectionsClient() {
	for k, c := range t.connsForClient {
		c.Close()
		delete(t.connsForClient, k)
	}
}

// Shutdown frees all connection and stop the listener
func (t *TCPNetwork) Shutdown() {
	if !atomic.CompareAndSwapInt32(&t.shutdownFlag, 0, 1) {
		return
	}

	//	stop accept routine
	if nil != t.listener {
		t.listener.Close()
	}

	//	close all connections
	t.DisconnectAllConnectionsClient()
	t.DisconnectAllConnectionsServer()
}

func (t *TCPNetwork) createConn(c net.Conn) *Connection {
	tcpConn := NewConnection(c, t.Conf.SendBufferSize, t)
	tcpConn.SetStreamProtocol(t.streamProtocol)
	return tcpConn
}

// ServeWithHandler process all events in the event queue and dispatch to the IEventHandler
func (t *TCPNetwork) ServeWithHandler(handler IEventHandler) {
	SERVE_LOOP:
	for {
		select {
		case evt, ok := <-t.eventQueue:
			{
				if !ok {
					//	channel closed or shutdown
					break SERVE_LOOP
				}

				t.handleEvent(evt, handler)
			}
		}
	}
}

func (t *TCPNetwork) acceptRoutine() {
	// after accept temporary failure, enter sleep and try again
	var tempDelay time.Duration

	for {
		tcpConn, err := t.listener.Accept()
		if err != nil {
			// check if the error is an temporary error
			if acceptErr, ok := err.(net.Error); ok && acceptErr.Temporary() {
				if 0 == tempDelay {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}

				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}

				logger.LogWarn("Accept error %s , retry after %d ms", acceptErr.Error(), tempDelay)
				time.Sleep(tempDelay)
				continue
			}

			logger.LogError("Accept routine quit.error: %s", err.Error())
			t.listener = nil
			return
		}

		//	process conn event
		connection := t.createConn(tcpConn)
		connection.SetReadTimeoutSec(t.readTimeoutSec)
		connection.From = kServerConn
		connection.Init()
		connection.Run()
	}
}

func (t *TCPNetwork) handleEvent(evt *ConnEvent, handler IEventHandler) {
	switch evt.EventType {
	case KConnEvent_Connected:
		{
			//	add to connection map
			connId := 0
			if kServerConn == evt.Conn.From {
				connId = t.connIdForServer + 1
				t.connIdForServer = connId
				t.connsForServer[connId] = evt.Conn
			} else {
				connId = t.connIdForClient + 1
				t.connIdForClient = connId
				t.connsForClient[connId] = evt.Conn
			}
			evt.Conn.ConnId = connId

			handler.OnConnected(evt)
		}
	case KConnEvent_Disconnected:
		{
			handler.OnDisconnected(evt)

			//	remove from connection map
			if kServerConn == evt.Conn.From {
				delete(t.connsForServer, evt.Conn.ConnId)
			} else {
				delete(t.connsForClient, evt.Conn.ConnId)
			}
		}
	case KConnEvent_Data:
		{
			handler.OnRecv(evt)
		}
	}
}
