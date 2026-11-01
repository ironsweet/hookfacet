package wire

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeFidelity(t *testing.T) {
	for _, tc := range []struct {
		name  string
		input []byte
	}{
		{name: "empty", input: []byte{}},
		{name: "one byte", input: []byte{0x01}},
		{name: "one kilo", input: make([]byte, 1024)},
		{name: "max - 1", input: make([]byte, MaxFrameSize-1)},
		{name: "max", input: make([]byte, MaxFrameSize)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			assert.NoError(t, Encode(&buf, tc.input))
			decoded, err := Decode(&buf)
			assert.NoError(t, err)
			assert.Equal(t, tc.input, decoded)
		})
	}
}

func TestEndian(t *testing.T) {
	var buf bytes.Buffer
	assert.NoError(t, Encode(&buf, []byte("AB")))
	assert.Equal(t, []byte{0x00, 0x00, 0x00, 0x02}, buf.Bytes()[:4])
}

type readWrapper struct {
	buf     *bytes.Buffer
	numRead int
}

func (r *readWrapper) Read(p []byte) (int, error) {
	n, err := r.buf.Read(p)
	r.numRead += n
	return n, err
}

var _ io.Reader = (*readWrapper)(nil)

func TestOverflow(t *testing.T) {
	input := make([]byte, MaxFrameSize+1)
	var buf bytes.Buffer
	err := Encode(&buf, input)
	assert.Error(t, err)
	assert.Equal(t, 0, buf.Len())

	var hdr [4]byte
	binary.BigEndian.PutUint32(hdr[:], MaxFrameSize+1)
	rw := &readWrapper{buf: bytes.NewBuffer(hdr[:])}
	_, err = Decode(rw)
	assert.Error(t, err)
	assert.Equal(t, 4, rw.numRead, "should have read the header before failing")
}

func TestEOFSemantics(t *testing.T) {
	var hdr [4]byte
	binary.BigEndian.PutUint32(hdr[:], 4)

	for _, tc := range []struct {
		name    string
		payload []byte
		assert  func(t *testing.T, err error)
	}{
		{name: "empty", payload: []byte{}, assert: func(t *testing.T, err error) {
			assert.ErrorIs(t, err, io.EOF)
		}},
		{name: "two bytes", payload: []byte{0x00, 0x00}, assert: func(t *testing.T, err error) {
			assert.ErrorIs(t, err, io.ErrUnexpectedEOF)
		}},
		{name: "bad length", payload: append(hdr[:], 0x01, 0x02), assert: func(t *testing.T, err error) {
			assert.ErrorIs(t, err, io.ErrUnexpectedEOF)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Decode(bytes.NewReader(tc.payload))
			tc.assert(t, err)
		})
	}
}

func TestMultiFrames(t *testing.T) {
	var buf bytes.Buffer
	assert.NoError(t, Encode(&buf, []byte("first")))
	assert.NoError(t, Encode(&buf, []byte("second")))

	decoded1, err := Decode(&buf)
	assert.NoError(t, err)
	assert.Equal(t, []byte("first"), decoded1)

	decoded2, err := Decode(&buf)
	assert.NoError(t, err)
	assert.Equal(t, []byte("second"), decoded2)
}
