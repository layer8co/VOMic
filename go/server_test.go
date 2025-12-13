package vomic

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"
)

type dummyWriter struct {
	buf   *bytes.Buffer
	fails bool
}

func (d dummyWriter) Write(b []byte) (int, error) {
	if d.fails {
		return 0, errors.New("error")
	}
	n, err := d.buf.Write(b)

	return n, err
}

func (d dummyWriter) RecieveContent() string {
	return d.buf.String()
}

// Testing the Pinger Function:
// 1. Context test
// 2. Writer Test
// 3. Timeout signal

// The Pinger respects the given context?
func TestContextPinger(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	reset := make(chan time.Duration, 8)

	timeout := make(chan bool, 1)

	go Pinger(ctx, io.Discard, reset, timeout)
	cancel()

	select {
	case _ = <-timeout:

	case <-time.After(200 * time.Millisecond):
		t.Fatal("Expected to timeout, but got nothing after waiting")
	}
}

// Does pinger actually pings?
func TestWriterPinger(t *testing.T) {
	ctx := context.Background()

	reset := make(chan time.Duration)

	timeout := make(chan bool)

	dummy := dummyWriter{
		buf: new(bytes.Buffer),
	}

	go Pinger(ctx, dummy, reset, timeout)

	reset <- time.Second

	time.Sleep(time.Second * 2)

	got := dummy.RecieveContent()

	if got == "" {
		t.Fatalf("Expected ping messages")
	}

}

// Does it sends the timeout signal currectly?
func TestTimeoutPinger(t *testing.T) {
	ctx := context.Background()

	reset := make(chan time.Duration, 8)
	reset <- time.Millisecond

	timeout := make(chan bool)

	dummy := dummyWriter{
		buf:   new(bytes.Buffer),
		fails: true,
	}

	go Pinger(ctx, dummy, reset, timeout)

	select {
	case _ = <-timeout:

	case <-time.After(200 * time.Millisecond):
		t.Fatal("Expected to timeout happen. Got nothing")
	}
}
