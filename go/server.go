package vomic

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type VOMic struct {
	// Pipe source is generally where the sound data will be written.
	PipeSource  io.Writer
	ControlPort uint16
	MediaPort   uint16
	InfoLogger  *log.Logger
	DebugLogger *log.Logger
}

// Blocks, And listens for new connection.
func (v VOMic) Listen() error {
	ctx, cancel := context.WithCancel(context.Background())
	// handling os signals gracefully is essential because otherwise deferring will not work.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	done := make(chan bool)

	go func() {
		sig := <-sigs
		v.DebugLogger.Printf("Recieved: %s signal\n", sig)
		v.InfoLogger.Println("Gracefully shutdown...")

		cancel()
		done <- true
	}()

	// Control socket
	c, err := net.Listen("tcp", fmt.Sprint(":", v.ControlPort))
	if err != nil {
		return err
	}

	v.InfoLogger.Println("🎤 VO-MIC Listening on TCP network: ", c.Addr())
	defer func() {
		_ = c.Close()
	}()

	var wg sync.WaitGroup

	// The main goroutine that faciliates the conneciton. Ideally, it never
	// stops.
	go func(wg *sync.WaitGroup) {
	MAIN:
		for {
			v.InfoLogger.Println("Waiting for new connection...")
			conn, err := c.Accept()

			// Accept() will immediately returns an error when
			// program exists. But it's error is generally not
			// critical.
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					v.DebugLogger.Println("TCP Socket closed.")
					return
				}

				v.InfoLogger.Println(err)
				continue
			}

			// Recieve the SessionInit packet from the client.
			buffer := make([]byte, SegmentSize)

			_, err = conn.Read(buffer)
			if err != nil {
				v.InfoLogger.Println(err)
				continue
			}

			var sInit SessionInit

			err = sInit.UnmarshalBinary(buffer)

			if err != nil {
				v.InfoLogger.Println("Handshake error: ", err)
				continue
			}

			v.InfoLogger.Printf("New client Connected: %s", sInit.Hostname)

			// Media socket
			m, err := net.ListenPacket("udp", fmt.Sprint(":", v.MediaPort))

			if err != nil {
				v.InfoLogger.Fatalf("Binding to udp socket %s failed: %s", m.LocalAddr().String(), err.Error())
				continue
			}

			// Pinger
			timeout := make(chan bool, 1)
			reset := make(chan time.Duration, 1)
			reset <- time.Second * 3

			go Pinger(ctx, conn, reset, timeout)

			// Listens for pinger timeout
			go func() {
				// Blocks until it gets the timeout sign from
				// Pinger. After that, the correspoing
				// connection loop will be breaked And the
				// program returns to the passive listening
				// state(the CONNECTION label).
				<-timeout
				cancel()
				_ = conn.Close()
				_ = m.Close()
			}()

			// Send Ack packet to the client
			var sessionAck SessionAck
			sessionAck.MediaPort = v.MediaPort

			bytes, err := sessionAck.MarshalBinary()
			if err != nil {
				cancel()
				v.InfoLogger.Fatalf("Handshake error: %s", err)
				continue
			}

			_, err = conn.Write(bytes)
			if err != nil {
				cancel()
				v.InfoLogger.Fatalf("Handshake error: %s", err)
				continue
			}

			writeCh := make(chan []byte, 200)
			errCh := make(chan error, 1)

			wg.Add(1)
			// This goroutine just writes anything that is gets from
			// the writeCh rapidly.
			go func() {
				defer wg.Done()
				for buf := range writeCh {
					if _, err := v.PipeSource.Write(buf); err != nil {
						errCh <- err
						fmt.Printf("pipeSource write error: %s", err)
						return
					}
				}
			}()

			for {
				buffer := make([]byte, 300)

				n, remote, err := m.ReadFrom(buffer)

				if err != nil {
					if errors.Is(err, net.ErrClosed) {
						v.DebugLogger.Println("UDP Socket closed.")
						break
					}
					v.InfoLogger.Println("udp socket error:", err.Error())
					continue
				}
				v.DebugLogger.Printf("read %d bytes from %s \n", n, remote.String())

				select {
				case <-timeout:
					cancel()
					close(writeCh)
					goto MAIN

				// Write everything you read from the
				// mediaSocket to pipeSource.
				case writeCh <- buffer[:n]:
				case err := <-errCh:
					v.InfoLogger.Println("Pipesource error: ", err)
				}
			}
		}
	}(&wg)

	<-done
	wg.Wait()

	return nil
}

// Pinger tries to make sure the connection is healthy and active. It sends a
// truthy value to timeout channel when it encounter with an error when writing
// on Writer. The reset channel act like a heartbeat of the process. The caller
// needs to keep advancing it after each network interaction to prevent
// unnecessary time-out packets be send. Zero value for reset channel defaults
// to 5 second duration. if it was set to default value,it sets the timer
// interval to 5 second.
func Pinger(ctx context.Context, w io.Writer, reset chan time.Duration, timeout chan bool) {
	var interval time.Duration

	select {
	case <-ctx.Done():
		timeout <- true
		return

	// assert the given interval, if anything was there.
	case interval = <-reset:

	default:

	}

	if interval <= 0 {
		interval = 5 * time.Second
	}

	timer := time.NewTimer(interval)

	defer func() {
		if !timer.Stop() {
			<-timer.C
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case newInternal := <-reset:
			if !timer.Stop() {
				<-timer.C
			}
			if newInternal > 0 {
				interval = newInternal
			}
		case <-timer.C:
			// Timer fired, so try to send something
			log.Println("Timer fired!")
			if _, err := w.Write([]byte("do not panic")); err != nil {
				close(timeout)
				return
			}
		}

		_ = timer.Reset(interval)

	}
}
