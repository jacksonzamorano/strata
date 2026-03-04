package hostio

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/jacksonzamorano/strata/internal/stdioipc"
)

type MessageType = HostMessageType
type Message = HostMessage

type IO struct {
	transport *stdioipc.IO[MessageType]
}

func NewIO(ctx context.Context, cancel context.CancelFunc, read io.ReadCloser, write io.Writer) *IO {
	return &IO{transport: stdioipc.NewIO[MessageType](ctx, cancel, read, write)}
}

func NewStdio(ctx context.Context, cancel context.CancelFunc) *IO {
	return NewIO(ctx, cancel, os.Stdin, os.Stdout)
}

func (io *IO) Done() <-chan struct{} {
	return io.transport.Done()
}

func (io *IO) Send(msg Message) bool {
	return io.transport.Send(stdioipc.Message[MessageType]{
		Id:      msg.Id,
		Type:    msg.Type,
		Payload: msg.Payload,
	})
}

func (io *IO) NewThread() *Thread {
	return &Thread{thread: io.transport.NewThread()}
}

type Thread struct {
	thread *stdioipc.Thread[MessageType]
}

func (t *Thread) ID() string {
	return t.thread.ID()
}

func (t *Thread) Send(typ MessageType, payload any) bool {
	return t.thread.Send(typ, payload)
}

type ReceivedEvent[T any] struct {
	Payload T
	Message Message
	Thread  *Thread
	Error   bool
}

func Receive[T any](io *IO, recvType MessageType) chan ReceivedEvent[T] {
	incoming := stdioipc.Receive[T](io.transport, recvType)
	output := make(chan ReceivedEvent[T])

	go func() {
		for {
			ev := <-incoming
			var thread *Thread
			if ev.Thread != nil {
				thread = &Thread{thread: ev.Thread}
			}
			output <- ReceivedEvent[T]{
				Payload: ev.Payload,
				Message: Message{
					Id:      ev.Message.Id,
					Type:    ev.Message.Type,
					Payload: ev.Message.Payload,
				},
				Thread: thread,
				Error:  ev.Error,
			}
			if ev.Error {
				return
			}
		}
	}()

	return output
}

func ReceiveOnce[T any](io *IO, timeout time.Duration, recvType MessageType) ReceivedEvent[T] {
	ev := stdioipc.ReceiveOnce[T](io.transport, timeout, recvType)
	var thread *Thread
	if ev.Thread != nil {
		thread = &Thread{thread: ev.Thread}
	}
	return ReceivedEvent[T]{
		Payload: ev.Payload,
		Message: Message{
			Id:      ev.Message.Id,
			Type:    ev.Message.Type,
			Payload: ev.Message.Payload,
		},
		Thread: thread,
		Error:  ev.Error,
	}
}

func SendAndReceive[T any](t *Thread, sendType MessageType, sendPayload any, recvType MessageType) (T, Message) {
	payload, msg := stdioipc.SendAndReceive[T](t.thread, sendType, sendPayload, recvType)
	return payload, Message{
		Id:      msg.Id,
		Type:    msg.Type,
		Payload: msg.Payload,
	}
}
