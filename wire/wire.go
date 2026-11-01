package wire

import "io"

const MaxFrameSize = 16 * 1024 * 1024 // 16 MiB

func Encode(w io.Writer, payload []byte) error {
	panic("niy")
}

func Decode(r io.Reader) ([]byte, error) {
	panic("niy")
}
