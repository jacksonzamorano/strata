package componentipc

import (
	"context"
	"io"
	"time"

	"github.com/jacksonzamorano/strata/internal/stdioipc"
)

type MessageType = ComponentMessageType
type Message = stdioipc.Message[MessageType]
type ReceivedEvent[T any] = stdioipc.ReceivedEvent[MessageType, T]
type Thread = stdioipc.Thread[MessageType]

type IO = stdioipc.IO[MessageType]

func NewIO(ctx context.Context, cancel context.CancelFunc, reader io.ReadCloser, writer io.Writer) *IO {
	return stdioipc.NewIO[MessageType](ctx, cancel, reader, writer)
}

func ReceiveOnce[T any](c *IO, timeout time.Duration, recvType MessageType) ReceivedEvent[T] {
	return stdioipc.ReceiveOnce[T](c, timeout, recvType)
}

func Receive[T any](c *IO, recvType MessageType) chan ReceivedEvent[T] {
	return stdioipc.Receive[T](c, recvType)
}

func SendAndReceive[T any](t *Thread, sendType MessageType, sendPayload any, recvType MessageType) (T, Message) {
	return stdioipc.SendAndReceive[T](t, sendType, sendPayload, recvType)
}
