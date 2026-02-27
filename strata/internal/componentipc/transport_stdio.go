package componentipc

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
)

type stdioTransport struct {
	read   chan ComponentMessage
	write  chan ComponentMessage
	ctx    context.Context
	cancel context.CancelFunc
}

func startStdioTransport(ctx context.Context, cancel context.CancelFunc, in io.ReadCloser, out io.Writer) *stdioTransport {
	st := &stdioTransport{
		read:   make(chan ComponentMessage, 64),
		write:  make(chan ComponentMessage, 64),
		ctx:    ctx,
		cancel: cancel,
	}

	go st.reader(in)
	go st.writer(out)

	return st
}

func (st *stdioTransport) writer(out io.Writer) {
	bw := bufio.NewWriter(out)
	for {
		select {
		case <-st.ctx.Done():
			return
		case msg := <-st.write:
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

func (st *stdioTransport) reader(in io.ReadCloser) {
	br := bufio.NewReader(in)
	go func() {
		<-st.ctx.Done()
		_ = in.Close()
	}()

	for {
		bytes, err := br.ReadBytes(0)
		if errors.Is(err, io.EOF) {
			st.cancel()
			return
		}
		if err != nil || len(bytes) == 0 {
			continue
		}

		var msg ComponentMessage
		if err := json.Unmarshal(bytes[:len(bytes)-1], &msg); err != nil {
			continue
		}
		st.read <- msg
	}
}

func (st *stdioTransport) send(msg ComponentMessage) {
	st.write <- msg
}
