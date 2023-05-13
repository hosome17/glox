package main

import (
	"glox"
	"os"
)

func main() {
	args := os.Args[1:]

	runtime := glox.NewRuntime()
	runtime.Run(args)
}
