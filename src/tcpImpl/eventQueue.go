package tcpImpl

import "tcp"

type EventQueue struct {

}

func NewEventQueue() *EventQueue {
    return &EventQueue{}
}

func (s *EventQueue) Push(*tcp.ConnEvent) {

}

func (s *EventQueue) Pop() *tcp.ConnEvent {
    return nil
}
