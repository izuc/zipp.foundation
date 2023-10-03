package tcp

import (
	"github.com/izuc/zipp.foundation/core/generics/event"
	"github.com/izuc/zipp.foundation/core/network"
)

type serverEvents struct {
	Start    *event.Event[*StartEvent]
	Shutdown *event.Event[*ShutdownEvent]
	Connect  *event.Event[*ConnectEvent]
	Error    *event.Event[error]
}

func newServerEvents() *serverEvents {
	return &serverEvents{
		Start:    event.New[*StartEvent](),
		Shutdown: event.New[*ShutdownEvent](),
		Connect:  event.New[*ConnectEvent](),
		Error:    event.New[error](),
	}
}

type StartEvent struct{}
type ShutdownEvent struct{}
type ConnectEvent struct {
	ManagedConnection *network.ManagedConnection
}
