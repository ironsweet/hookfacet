package wire

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const MaxFrameSize = 16 * 1024 * 1024 // 16 MiB

var ErrFrameTooLarge = errors.New("frame too large")

func Encode(w io.Writer, payload []byte) error {
	if len(payload) > MaxFrameSize {
		return ErrFrameTooLarge
	}

	var hdr [4]byte
	binary.BigEndian.PutUint32(hdr[:], uint32(len(payload)))
	_, err := w.Write(hdr[:])
	if err != nil {
		return fmt.Errorf("write frame header: %w", err)
	}

	_, err = w.Write(payload)
	if err != nil {
		return fmt.Errorf("write frame payload: %w", err)
	}

	return nil
}

func Decode(r io.Reader) ([]byte, error) {
	var hdr [4]byte
	_, err := io.ReadFull(r, hdr[:])
	if err != nil {
		return nil, fmt.Errorf("read frame header: %w", err)
	}

	size := binary.BigEndian.Uint32(hdr[:])
	if size == 0 {
		return []byte{}, nil
	}
	if size > MaxFrameSize {
		return nil, ErrFrameTooLarge
	}

	payload := make([]byte, size)
	_, err = io.ReadFull(r, payload)
	if err != nil {
		if errors.Is(err, io.EOF) {
			err = io.ErrUnexpectedEOF
		}
		return nil, fmt.Errorf("read frame payload: %w", err)
	}

	return payload, nil
}
