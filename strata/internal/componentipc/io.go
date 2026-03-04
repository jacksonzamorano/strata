package componentipc

import (
	"context"
	"io"
	"time"

	"github.com/jacksonzamorano/strata/internal/stdioipc"
)

type IOMessage = stdioipc.Message[MessageType]
type IO = stdioipc.IO[MessageType]
type Thread = stdioipc.Thread[MessageType]
type ReceivedEvent[T any] = stdioipc.ReceivedEvent[MessageType, T]

func NewIO(ctx context.Context, cancel context.CancelFunc, read io.ReadCloser, write io.Writer) *IO {
	return stdioipc.NewIO[MessageType](ctx, cancel, read, write)
}

func Receive[T any](c *IO, recvType MessageType) chan ReceivedEvent[T] {
	return stdioipc.Receive[T](c, recvType)
}

func ReceiveOnce[T any](c *IO, timeout time.Duration, recvType MessageType) ReceivedEvent[T] {
	return stdioipc.ReceiveOnce[T](c, timeout, recvType)
}

func SendAndReceive[T any](t *Thread, sendType MessageType, sendPayload any, recvType MessageType) (T, IOMessage) {
	return stdioipc.SendAndReceive[T](t, sendType, sendPayload, recvType)
}
