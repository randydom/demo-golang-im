package tcpImpl

import "tcp"

type Unpack struct {

}

func NewUnpack() *Unpack {
    return &Unpack{}
}

func (s *EventHandler) Unpack(*tcp.Connection, []byte) ([]byte, error) {
    return nil, nil
}
