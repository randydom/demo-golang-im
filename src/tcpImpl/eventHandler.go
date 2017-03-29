package tcpImpl

import "tcp"

type EventHandler struct {

}

func NewEventHandler() *EventHandler {
    return &EventHandler{}
}

func (s *EventHandler) OnConnected(*tcp.ConnEvent) {

}

func (s *EventHandler) OnDisconnected(*tcp.ConnEvent) {

}

func (s *EventHandler) OnRecv(*tcp.ConnEvent) {

}
