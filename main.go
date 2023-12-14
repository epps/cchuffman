package main

import (
	"flag"
	"log"
)

func main() {
	input := flag.String("input", "les-mis.txt", "the input file to encode")
	output := flag.String("output", "output.txt", "the output filepath")
	decompress := flag.Bool("decompress", false, "treat the input file as compressed")

	flag.Parse()

	if !*decompress {
		encoder := NewHuffmanEncoder(*input, *output)

		if err := encoder.Encode(); err != nil {
			log.Fatalf("failed to compress %s: %v", *input, err)
		}
	} else {
		decoder := NewHuffmanDecoder(*input, *output)

		if err := decoder.Decode(); err != nil {
			log.Fatalf("failed to decompress %s: %v", *input, err)
		}
	}
}
