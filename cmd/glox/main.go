package main

import (
	"glox"
	"os"
)

func main() {
	args := os.Args[1:]

	g := glox.NewGlox()
	g.Run(args)
}
