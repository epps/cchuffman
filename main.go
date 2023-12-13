package main

import (
	"flag"
	"log"
)

func main() {
	input := flag.String("input", "les-mis.txt", "the input file to encode")
	output := flag.String("output", "output.txt", "the output filepath")

	flag.Parse()

	encoder := NewHuffmanEncoder(*input, *output)

	if err := encoder.Encode(); err != nil {
		log.Fatalf("failed to compress %s: %v", *input, err)
	}
}
