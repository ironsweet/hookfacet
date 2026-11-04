package main

import (
	"io"

	"github.com/ironsweet/hookfacet"
)

type Harness interface {
	Read(r io.Reader) (hookfacet.Event, error)
	Write(w io.Writer, ev hookfacet.Decision) error
}
