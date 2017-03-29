package tcp

type IEventQueue interface {
	Push(*ConnEvent)
	Pop() *ConnEvent
}

type IStreamProtocol interface {
	//	Init
	Init()
	//	get the header length of the stream
	GetHeaderLength() uint32
	//	read the header length of the stream
	UnserializeHeader([]byte) uint32
	//	format header
	SerializeHeader([]byte) []byte
}

type IEventHandler interface {
	OnConnected(evt *ConnEvent)
	OnDisconnected(evt *ConnEvent)
	OnRecv(evt *ConnEvent)
}

type IUnpacker interface {
	Unpack(*Connection, []byte) ([]byte, error)
}
