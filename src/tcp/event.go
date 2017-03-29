package tcp

// ConnEvent represents a event occurs on a connection, such as connected, disconnected or data arrived
type ConnEvent struct {
    EventType int
    Conn      *Connection
    Data      []byte
    Extra     interface{}
}

func NewConnEvent(et int, c *Connection, d []byte) *ConnEvent {
    return &ConnEvent{
        EventType: et,
        Conn:      c,
        Data:      d,
    }
}