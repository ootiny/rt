package main

import (
	"log"

	"github.com/ootiny/rt/builder"
)

func main() {
	if err := builder.Output(); err != nil {
		log.Panicf("Failed to output: %v", err)
	}
}
