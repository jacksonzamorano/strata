package stdioipc

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
)

type Transport[MT ~string] struct {
	read   chan Message[MT]
	write  chan Message[MT]
	ctx    context.Context
	cancel context.CancelFunc
}

func NewTransport[MT ~string](ctx context.Context, cancel context.CancelFunc, in io.ReadCloser, out io.Writer) *Transport[MT] {
	transport := &Transport[MT]{
		read:   make(chan Message[MT], 64),
		write:  make(chan Message[MT], 64),
		ctx:    ctx,
		cancel: cancel,
	}

	go transport.reader(in)
	go transport.writer(out)

	return transport
}

func (t *Transport[MT]) Read() <-chan Message[MT] {
	return t.read
}

func (t *Transport[MT]) Done() <-chan struct{} {
	return t.ctx.Done()
}

func (t *Transport[MT]) Send(msg Message[MT]) bool {
	select {
	case <-t.ctx.Done():
		return false
	case t.write <- msg:
		return true
	}
}

func (t *Transport[MT]) writer(out io.Writer) {
	bw := bufio.NewWriter(out)
	for {
		select {
		case <-t.ctx.Done():
			return
		case msg := <-t.write:
			payload, err := json.Marshal(msg)
			if err != nil {
				continue
			}
			if _, err := bw.Write(payload); err != nil {
				continue
			}
			if err := bw.WriteByte(0); err != nil {
				continue
			}
			_ = bw.Flush()
		}
	}
}

func (t *Transport[MT]) reader(in io.ReadCloser) {
	br := bufio.NewReader(in)
	go func() {
		<-t.ctx.Done()
		_ = in.Close()
	}()

	for {
		bytes, err := br.ReadBytes(0)
		if errors.Is(err, io.EOF) {
			t.cancel()
			return
		}
		if err != nil || len(bytes) == 0 {
			continue
		}

		var msg Message[MT]
		if err := json.Unmarshal(bytes[:len(bytes)-1], &msg); err != nil {
			continue
		}

		select {
		case <-t.ctx.Done():
			return
		case t.read <- msg:
		}
	}
}
