package component

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
)

type StdioTransport struct {
	read   chan ComponentMessage
	write  chan ComponentMessage
	ctx    context.Context
	cancel context.CancelFunc
}

func StartStdioTransport(in io.ReadCloser, out io.Writer) StdioTransport {

	ctx, cancel := context.WithCancel(context.Background())

	read := make(chan ComponentMessage, 64)
	write := make(chan ComponentMessage, 64)

	st := StdioTransport{
		read,
		write,
		ctx,
		cancel,
	}

	go st.reader(in)
	go st.writer(out)

	return st
}

func (st *StdioTransport) writer(out io.Writer) {
	bw := bufio.NewWriter(out)
	for {
		select {
		case <-st.ctx.Done():
			return
		case l := <-st.write:
			msg_bytes, _ := json.Marshal(l)
			bw.Write(msg_bytes)
			bw.WriteByte(0)
			bw.Flush()
		}
	}
}
func (st *StdioTransport) reader(in io.ReadCloser) {
	br := bufio.NewReader(in)
	go func() {
		<-st.ctx.Done()
		in.Close()
	}()
	for {
		bytes, err := br.ReadBytes(0)
		if errors.Is(err, io.EOF) {
			return
		}
		var msg ComponentMessage
		err = json.Unmarshal(bytes[:len(bytes)-1], &msg)
		if err != nil {
			continue
		}
		st.read <- msg
	}
}

func (st *StdioTransport) Send(msg ComponentMessage) {
	st.write <- msg
}
func (st *StdioTransport) Read() ComponentMessage {
	return <- st.read
}
func (st *StdioTransport) Cancel() {
	st.cancel()
}
