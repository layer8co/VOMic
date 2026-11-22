package vomic

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"sync"
	"testing"
	"time"
)

type dummyWriter struct {
	buf   *bytes.Buffer
	fails bool

	// mutex protects the "buf".
	mu sync.Mutex
}

func (d *dummyWriter) Write(b []byte) (int, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.fails {
		return 0, errors.New("error")
	}

	return d.buf.Write(b)
}

func (d *dummyWriter) RecieveContent() string {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.buf.String()
}

// Testing the Pinger Function:
// 1. Context test
// 2. Writer Test
// 3. Timeout signal

// The Pinger respects the given context?
func TestContextPinger(t *testing.T) {
	t.Parallel()

	v := newTestVOMic()
	ctx, cancel := context.WithCancel(context.Background())

	reset := make(chan time.Duration, 8)

	go v.Pinger(ctx, io.Discard, reset)
	cancel()

	select {
	case <-ctx.Done():

	case <-time.After(200 * time.Millisecond):
		t.Fatal("Expected to timeout, but got nothing after waiting for 200 miliseconds!")
	}
}

// Does pinger actually pings?
func TestWriterPinger(t *testing.T) {
	t.Parallel()

	v := newTestVOMic()

	ctx := context.Background()

	reset := make(chan time.Duration, 1)

	reset <- time.Second * 1

	dummy := dummyWriter{
		buf: new(bytes.Buffer),
	}

	go v.Pinger(ctx, &dummy, reset)

	time.Sleep(time.Second * 2)

	got := dummy.RecieveContent()

	if got == "" {
		t.Fatalf("Expected ping messages")
	}

}

// Does it sends the timeout signal currectly?
func TestTimeoutPinger(t *testing.T) {
	t.Parallel()

	v := newTestVOMic()
	ctx, cancel := context.WithCancel(context.Background())

	reset := make(chan time.Duration, 8)
	reset <- time.Millisecond

	dummy := dummyWriter{
		buf:   new(bytes.Buffer),
		fails: true,
	}

	go func() {
		err := v.Pinger(ctx, &dummy, reset)
		if err != nil {
			cancel()
		}
	}()

	select {
	case <-ctx.Done():
		// Expecting cancelation, as a side effect of the error in goroutine
		// above.

	case <-time.After(200 * time.Millisecond):
		t.Fatal("Expected to recieve a context cancelation. Got nothing.")
	}
}

func newTestVOMic() VOMic {
	return VOMic{
		DebugLogger: log.New(io.Discard, "", log.Default().Flags()),
	}
}
