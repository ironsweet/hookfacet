package main

import (
	"os"
)

func main() {
	code := run(os.Stdin, os.Stdout, os.Getenv)
	os.Exit(code)
}
