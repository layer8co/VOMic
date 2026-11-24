package vomic

import (
	"bytes"
	"encoding/binary"
	"errors"
	"strings"
)

type PactlConf struct {
	SampleRate int32

	// The pipeSource filename
	PipePath string

	// It's the name that is displayed to the users when they want to select
	// this microphone source.
	SourceName string

	// The list of availble formats are listed in the url below (We are
	// using signed little endian 16 bit):
	// https://docs.pipewire.org/page_man_pipewire-pulse-modules_7.html
	Format PWFormat

	// Mono or Stereo.
	Channel PWChannel
}

type PWFormat string

const (
	Signed_le_16 PWFormat = "s16le"
	Signed_be_16 PWFormat = "s16be"
)

type PWChannel uint8

const (
	Mono   PWChannel = 1
	Stereo PWChannel = 2
)

const (
	// SessionInit should not be bigger then this.
	SegmentSize = 1 << 6
)

// SessionInit is a kind of message that we are expecting as a first payload.
// It used to convey some options from client. Think of it as a application
// level handshake.
type SessionInit struct {
	Flags    PWChannel
	Hostname string
}

// SessionAck is a kind of message that informs the client about the MediaPort
// we are listening to.
type SessionAck struct {
	MediaPort uint16
}

func (pc PWChannel) HasFlag(flag PWChannel) bool {
	// Bitwise AND
	return flag&pc != 0
}

type Bitmask uint8

const (
	// Mono channel or Stereo?
	FlagStereo PWChannel = 1 << 0
)

func (s *SessionInit) UnmarshalBinary(si []byte) error {
	b := bytes.NewBuffer(si)

	err := binary.Read(b, binary.BigEndian, &s.Flags)

	if err != nil {
		return err
	}

	s.Hostname, err = b.ReadString(0)
	if err != nil {
		return err
	}

	// ReadString reads until null character and includes it, here we strip
	// it.
	s.Hostname = strings.TrimRight(s.Hostname, "\x00")
	if s.Hostname == "" {
		return errors.New("a client with empty hostname!")
	}

	return nil
}

func (s SessionInit) MarshalBinary() ([]byte, error) {
	buffer := new(bytes.Buffer)

	cap := 1 + len(s.Hostname)
	buffer.Grow(cap)

	err := binary.Write(buffer, binary.BigEndian, s.Flags)
	if err != nil {
		return nil, err
	}

	_, err = buffer.WriteString(s.Hostname)
	if err != nil {
		return nil, err
	}

	err = buffer.WriteByte(0)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (sa SessionAck) MarshalBinary() ([]byte, error) {
	buffer := new(bytes.Buffer)

	cap := binary.Size(sa.MediaPort)

	buffer.Grow(cap)

	err := binary.Write(buffer, binary.BigEndian, sa.MediaPort)

	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
