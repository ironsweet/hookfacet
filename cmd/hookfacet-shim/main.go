package main

import (
	"io"
	"os"
)

func main() {
	code := run(os.Stdin, os.Stdout, os.Getenv)
	os.Exit(code)
}

func run(r io.Reader, w io.Writer, getenv func(string) string) int {
	panic("niy")
}
