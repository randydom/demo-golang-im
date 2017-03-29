package tcp

import (
	"fmt"
	"net"
	"sync/atomic"
	"time"
	"runtime/debug"
	"logger"
    "io"
)

const (
	kConnStatus_None = iota
	kConnStatus_Connected
	kConnStatus_Disconnected
)

// All connection event
const (
	KConnEvent_None = iota
	KConnEvent_Connected
	KConnEvent_Disconnected
	KConnEvent_Data
	KConnEvent_Pb
	KConnEvent_Close
	KConnEvent_Total
)

const (
	kConnConf_DefaultSendTimeoutSec = 5
	kConnConf_MaxReadBufferLength   = 0xffff // 0xffff
)

// Send method flag
const (
	KConnFlag_CopySendBuffer = 1 << iota // do not copy the send buffer
	KConnFlag_NoHeader                   // do not append stream header
)

// Connection is a wrap for net.Conn and process read and write task of the conn
// When event occurs, it will call the eventQueue to dispatch event
type Connection struct {
	ConnId              int
	From                int
	conn                net.Conn
	status              int32
	sendMsgQueue        chan *SendTask
	sendTimeoutSec      int
	eventQueue          IEventQueue
	streamProtocol      IStreamProtocol
	maxReadBufferLength int
	extra               interface{}
	readTimeoutSec      int
	fnSyncExecute       FuncSyncExecute
	unpacker            IUnpacker
	disableSend         int32
	localAddr           string
	remoteAddr          string
}

func NewConnection(c net.Conn, sendBufferSize int, eq IEventQueue) *Connection {
	return &Connection{
		ConnId:              0,
		conn:                c,
		status:              kConnStatus_None,
		sendMsgQueue:        make(chan *SendTask, sendBufferSize),
		sendTimeoutSec:      kConnConf_DefaultSendTimeoutSec,
		maxReadBufferLength: kConnConf_MaxReadBufferLength,
		eventQueue:          eq,
	}
}

func (c *Connection) Init() {
	c.localAddr = c.conn.LocalAddr().String()
	c.remoteAddr = c.conn.RemoteAddr().String()
}

//	directly close, packages in queue will not be sent
func (c *Connection) close() {
	//	set the disconnected status, use atomic operation to avoid close twice
	if atomic.CompareAndSwapInt32(&c.status, kConnStatus_Connected, kConnStatus_Disconnected) {
		c.conn.Close()
	}
}

// Close the connection, routine safe, send task in the queue will be sent before closing the connection
func (c *Connection) Close() {
	if atomic.LoadInt32(&c.status) != kConnStatus_Connected {
		return
	}

	select {
	case c.sendMsgQueue <- nil:
		{
			//	nothing
		}
	case <-time.After(time.Duration(c.sendTimeoutSec) * time.Second):
		{
			//	timeout, close the connection
			c.close()
		}
	}

	//	disable send
	atomic.StoreInt32(&c.disableSend, 1)
}

//	When don't need conection to send any thing, free it, DO NOT call it on multi routines
func (c *Connection) Free() {
	if nil != c.sendMsgQueue {
		close(c.sendMsgQueue)
		c.sendMsgQueue = nil
	}
}

func (c *Connection) SyncExecuteEvent(evt *ConnEvent) bool {
	if nil == c.fnSyncExecute {
		return false
	}

	return c.fnSyncExecute(evt)
}

func (c *Connection) PushEvent(et int, d []byte) {
	//	this is for sync execute
	evt := NewConnEvent(et, c, d)
	if c.SyncExecuteEvent(evt) {
		return
	}

	if nil == c.eventQueue {
		panic("Nil event queue")
		return
	}
	c.eventQueue.Push(evt)
}

// SetSyncExecuteFunc , you can set a callback that you can synchoronously process the event in every connection's event routine
// If the callback function return true, the event will not be dispatched
func (c *Connection) SetSyncExecuteFunc(fn FuncSyncExecute) FuncSyncExecute {
	prevFn := c.fnSyncExecute
	c.fnSyncExecute = fn
	return prevFn
}

// GetStatus get the connection's status
func (c *Connection) GetStatus() int32 {
	return c.status
}

func (c *Connection) SetStatus(stat int) {
	c.status = int32(stat)
}

// GetConnId get the connection's id
func (c *Connection) GetConnId() int {
	return c.ConnId
}

// SetConnId set the connection's id
func (c *Connection) SetConnId(id int) {
	c.ConnId = id
}

// GetConn get the raw net.Conn interface
func (c *Connection) GetConn() net.Conn {
	return c.conn
}

// GetExtra get the extra you set
func (c *Connection) GetExtra() interface{} {
	return c.extra
}

// SetExtra set the extra you need
func (c *Connection) SetExtra(extra interface{}) {
	c.extra = extra
}

// SetReadTimeoutSec set the read deadline for the connection
func (c *Connection) SetReadTimeoutSec(sec int) {
	c.readTimeoutSec = sec
}

//  GetReadTimeoutSec get the read deadline for the connection
func (c *Connection) GetReadTimeoutSec() int {
	return c.readTimeoutSec
}

// GetRemoteAddress return the remote address of the connection
func (c *Connection) GetRemoteAddress() string {
	return c.remoteAddr
}

// GetLocalAddress return the local address of the connection
func (c *Connection) GetLocalAddress() string {
	return c.localAddr
}

// SetUnpacker you can set a custom binary stream unpacker on the connection
func (c *Connection) SetUnpacker(unpacker IUnpacker) {
	c.unpacker = unpacker
}

// GetUnpacker you can get the unpacker you set
func (c *Connection) GetUnpacker() IUnpacker {
	return c.unpacker
}

func (c *Connection) SetStreamProtocol(sp IStreamProtocol) {
	c.streamProtocol = sp
}

func (c *Connection) SendRaw(task *SendTask) error {
	if atomic.LoadInt32(&c.disableSend) != 0 {
		return ErrConnIsClosed
	}
	if atomic.LoadInt32(&c.status) != kConnStatus_Connected {
		return ErrConnIsClosed
	}

	select {
	case c.sendMsgQueue <- task:
		{
			//	nothing
		}
	case <-time.After(time.Duration(c.sendTimeoutSec) * time.Second):
		{
			//	timeout, close the connection
			logger.LogError("Send to peer %s timeout, close connection!", c.GetRemoteAddress())
			c.close()
			return ErrConnSendTimeout
		}
	}

	return nil
}

// ApplyReadDealine
func (c *Connection) ApplyReadDeadline() {
	if 0 != c.readTimeoutSec {
		c.conn.SetReadDeadline(time.Now().Add(time.Duration(c.readTimeoutSec) * time.Second))
	}
}

// ResetReadDeadline
func (c *Connection) ResetReadDeadline() {
	c.conn.SetReadDeadline(time.Time{})
}

// Send the buffer
func (c *Connection) Send(msg []byte, f int64) error {
	task := &SendTask{
		data: msg,
		flag: f,
	}
	buf := msg

	//	copy send buffer
	if 0 != f&KConnFlag_CopySendBuffer {
		msgCopy := make([]byte, len(msg))
		copy(msgCopy, msg)
		buf = msgCopy
		task.data = buf
	}

	return c.SendRaw(task)
}

//	run a routine to process the connection
func (c *Connection) Run() {
	go c.RoutineMain()
}

func (c *Connection) RoutineMain() {
	defer func() {
		//	routine end
		e := recover()
		if e != nil {
			logger.LogFatal("Read routine panic %v, stack:", e)
			stackInfo := debug.Stack()
			logger.LogFatal(string(stackInfo))
		}

		//	close the connection
		c.close()

		//	free channel
		//	FIXED : consumers need free it, not producer

		//	post event
		c.PushEvent(KConnEvent_Disconnected, nil)
	}()

	if nil == c.streamProtocol {
		panic("Nil stream protocol")
		return
	}
	c.streamProtocol.Init()

	//	connected
	c.PushEvent(KConnEvent_Connected, nil)
	atomic.StoreInt32(&c.status, kConnStatus_Connected)

	go c.RoutineSend()
	err := c.RoutineRead()
	if nil != err {
		logger.LogError("Read routine quit with error: %v", err)
	}
}

func (c *Connection) RoutineSend() error {
	var err error

	defer func() {
		if nil != err {
			logger.LogError("Send routine quit with error: %v", err)
		}
		e := recover()
		if nil != e {
			//	panic
			logger.LogFatal("Send routine panic %v, stack: ", e)
			stackInfo := debug.Stack()
			logger.LogFatal(string(stackInfo))
		}
	}()

	for {
		select {
		case evt, ok := <-c.sendMsgQueue:
			{
				if !ok {
					//	channel closed, quit
					return nil
				}

				if nil == evt {
					c.close()
					return nil
				}

				if 0 == evt.flag&KConnFlag_NoHeader {
					headerBytes := c.streamProtocol.SerializeHeader(evt.data)
					if nil != headerBytes {
						//	write header first
						if len(headerBytes) != 0 {
							_, err = c.conn.Write(headerBytes)
							if err != nil {
								return err
							}
						}
					} else {
						//	invalid packet
						panic("Failed to serialize header")
						break
					}
				}

				_, err = c.conn.Write(evt.data)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (c *Connection) RoutineRead() error {
	//	default buffer
	buf := make([]byte, c.maxReadBufferLength)
	var msg []byte
	var err error

	for {
		if nil == c.unpacker {
			msg, err = c.Unpack(buf)
		} else {
			msg, err = c.unpacker.Unpack(c, buf)
		}
		if err != nil {
            if err == io.EOF {
                break
            } else {
                return err
            }
		}
		if nil == msg {
			// Empty pstream
			continue
		}

		//	only push event when the connection is connected
		if atomic.LoadInt32(&c.status) == kConnStatus_Connected {
			c.PushEvent(KConnEvent_Data, msg)
		}
	}

	return nil
}

func (c *Connection) Unpack(buf []byte) ([]byte, error) {
	//	read head
	c.ApplyReadDeadline()
	headerLength := int(c.streamProtocol.GetHeaderLength())
	if headerLength > len(buf) {
		return nil, fmt.Errorf("Header length %d > buffer length %d", headerLength, len(buf))
	}
	headBuf := buf[:headerLength]
	_, err := c.conn.Read(headBuf)
	if err != nil {
		return nil, err
	}

	//	check length
	packetLength := c.streamProtocol.UnserializeHeader(headBuf)
	if packetLength > uint32(c.maxReadBufferLength) ||
		packetLength < c.streamProtocol.GetHeaderLength() {
		return nil, fmt.Errorf("Invalid stream length %d", packetLength)
	}
	if packetLength == c.streamProtocol.GetHeaderLength() {
		// Empty stream ?
		logger.LogFatal("Invalid stream length equal to header length")
		return nil, nil
	}

	//	read body
	c.ApplyReadDeadline()
	bodyLength := packetLength - c.streamProtocol.GetHeaderLength()
	_, err = c.conn.Read(buf[:bodyLength])
	if err != nil {
		return nil, err
	}

	//	ok
	msg := make([]byte, bodyLength)
	copy(msg, buf[:bodyLength])
	c.ResetReadDeadline()

	return msg, nil
}
