package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/jacksonzamorano/tasks/tasklib/component"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--client" {
		runClientMode()
		return
	}

	if err := runServerMode(); err != nil {
		fmt.Fprintf(os.Stderr, "server_error:%v\n", err)
		os.Exit(1)
	}
}

func runServerMode() error {
	executablePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve executable path: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	clientCmd := exec.CommandContext(ctx, executablePath, "--client")
	clientStdout, err := clientCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("client stdout pipe: %w", err)
	}
	clientStdin, err := clientCmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("client stdin pipe: %w", err)
	}
	clientStderr, err := clientCmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("client stderr pipe: %w", err)
	}

	if err := clientCmd.Start(); err != nil {
		return fmt.Errorf("start client mode: %w", err)
	}

	clientStderrDone := make(chan struct{})
	go func() {
		defer close(clientStderrDone)
		_, _ = io.Copy(os.Stderr, clientStderr)
	}()

	transport := component.Start(clientStdout, clientStdin)
	fmt.Println("server_started")

	for idx, expectedPayload := range []string{"hello-from-client-1", "hello-from-client-2"} {
		msg := transport.Read()
		fmt.Printf("server_received[%d]:%s\n", idx+1, msg.Payload)
		if msg.Payload != expectedPayload {
			return fmt.Errorf("unexpected client payload at index %d: got %q want %q", idx+1, msg.Payload, expectedPayload)
		}

		ackPayload := "ack:" + msg.Payload
		transport.Send(component.ComponentMessage{
			Type:    msg.Type,
			Payload: ackPayload,
		})
		fmt.Printf("server_sent[%d]:%s\n", idx+1, ackPayload)
	}

	// Give the writer goroutine a small window to flush bytes before process exit.
	time.Sleep(120 * time.Millisecond)
	transport.Cancel()
	_ = clientStdin.Close()

	waitErr := clientCmd.Wait()
	<-clientStderrDone
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("client timed out: %w", ctx.Err())
	}
	if waitErr != nil {
		return fmt.Errorf("client exit: %w", waitErr)
	}

	fmt.Println("server_exit")
	return nil
}

func runClientMode() {
	transport := component.Start(os.Stdin, os.Stdout)

	for _, payload := range []string{"hello-from-client-1", "hello-from-client-2"} {
		transport.Send(component.ComponentMessage{
			Type:    component.ComponentMessageTypeInitialize,
			Payload: payload,
		})

		resp := transport.Read()
		expectedAck := "ack:" + payload
		if resp.Payload != expectedAck {
			fmt.Fprintf(os.Stderr, "client_unexpected_payload:%q expected:%q\n", resp.Payload, expectedAck)
			os.Exit(1)
		}
	}

	transport.Cancel()
}
