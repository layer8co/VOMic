package vomic

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
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
func (v VOMic) ListenAndServe(ctx context.Context) error {
	// Control socket
	c, err := net.Listen("tcp", fmt.Sprint(":", v.ControlPort))
	if err != nil {
		return err
	}
	defer c.Close()

	go func() {
		<-ctx.Done()
		_ = c.Close()
	}()

	v.InfoLogger.Println("🎤 VO-MIC Listening on TCP network: ", c.Addr())

	for {
		v.InfoLogger.Println("Waiting for new connection...")

		conn, err := c.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}

			v.InfoLogger.Println("Accept error:", err)
			continue
		}

		// This must be block.
		if err := v.handleSession(ctx, conn); err != nil {
			v.InfoLogger.Println("Session ended. Cause:", err.Error())
		}

		if ctx.Err() != nil {
			return nil
		}
	}
}

func (v VOMic) handleSession(ctx context.Context, conn net.Conn) error {
	sessionCtx, sessionCancel := context.WithCancel(ctx)
	defer sessionCancel()

	// Recieve the SessionInit packet from the client.
	buffer := make([]byte, SegmentSize)

	_, err := conn.Read(buffer)
	if err != nil {
		return fmt.Errorf("SessionInit Handshake error: %w", err)
	}
	var sInit SessionInit

	err = sInit.UnmarshalBinary(buffer)

	if err != nil {
		return fmt.Errorf("SessionInit Handshake error: %w", err)
	}

	v.InfoLogger.Printf("New client has been connected: %s", sInit.Hostname)

	// Media socket
	m, err := net.ListenPacket("udp", fmt.Sprint(":", v.MediaPort))

	if err != nil {
		err := fmt.Sprintf("binding to udp socket failed: %w", err)
		return errors.New(err)
	}
	defer m.Close()

	go func() {
		<-sessionCtx.Done()
		m.Close()
	}()

	// Setup Pinger.
	errCh := make(chan error, 2)

	reset := make(chan time.Duration, 1)
	reset <- 3 * time.Second

	go func() {
		if err := Pinger(sessionCtx, conn, reset); err != nil {
			select {
			case errCh <- err:
			default:
			}
			sessionCancel()
		}
	}()

	// Send Ack packet to the client
	var sessionAck SessionAck
	sessionAck.MediaPort = v.MediaPort

	bytes, err := sessionAck.MarshalBinary()
	if err != nil {
		err := fmt.Sprintf("SessionAck marshaler failed: %w", err)
		return errors.New(err)
	}

	_, err = conn.Write(bytes)
	if err != nil {
		err := fmt.Sprintf("Writing SessionAck failed: %w", err)
		return errors.New(err)
	}

	writeCh := make(chan []byte, 200)

	// This goroutine just writes anything that is gets from
	// the writeCh rapidly.
	go func() {
		for {
			select {
			case <-sessionCtx.Done():
				return

			case buf, ok := <-writeCh:
				if !ok {
					return
				}

				if _, err := v.PipeSource.Write(buf); err != nil {
					errCh <- fmt.Errorf("PipeSource write failed: %w", err)

					sessionCancel()
					return
				}
			}
		}
	}()

	for {
		buffer := make([]byte, 200)

		n, remote, err := m.ReadFrom(buffer)

		if err != nil {
			if sessionCtx.Err() != nil {
				select {
				case workerErr := <-errCh:
					return fmt.Errorf("session worker failed: %w", workerErr)
				default:
					return nil
				}
			}

			return fmt.Errorf("UDP read error: %w", err)
		}

		v.DebugLogger.Printf("read %d bytes from %s", n, remote)

		select {
		case <-sessionCtx.Done():
			select {
			case workerErr := <-errCh:
				return fmt.Errorf("session worker failed: %w", workerErr)
			default:
				return nil
			}

		case writeCh <- buffer[:n]:
		}
	}
}

// Pinger tries to make sure the connection is healthy and active. It sends a
// truthy value to timeout channel when it encounter with an error when writing
// on Writer. The reset channel act like a heartbeat of the process. The caller
// needs to keep advancing it after each network interaction to prevent
// unnecessary time-out packets be send. Zero value for reset channel defaults
// to 5 second duration. if it was set to default value,it sets the timer
// interval to 5 second.
func Pinger(
	ctx context.Context,
	w io.Writer,
	reset <-chan time.Duration,
) error {
	interval := 5 * time.Second

	timer := time.NewTimer(interval)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Pinger context done")
			return nil

		case newInterval, ok := <-reset:
			if !ok {
				return nil
			}

			if newInterval > 0 {
				interval = newInterval
			}

			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}

			timer.Reset(interval)

		case <-timer.C:
			log.Println("Pinger pinging")
			if _, err := w.Write([]byte("do not panic")); err != nil {
				return fmt.Errorf("heartbeat write failed: %w", err)
			}

			timer.Reset(interval)
		}
	}
}
