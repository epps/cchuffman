package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	input := flag.String("input", "les-mis.txt", "the input file to encode")
	output := flag.String("output", "compressed.txt", "the output filepath")

	flag.Parse()

	ft := NewFrequencyTable(*input)

	if err := ft.Populate(); err != nil {
		log.Fatalf("error populating frequency table: %v\n", err)
	}

	fmt.Printf("Logging character frequencies for %s:\n", *input)
	if err := ft.Log(); err != nil {
		log.Fatalf("error logging frequency table: %v", err)
	}

	pq := NewPriorityQueue(ft.ToList())

	root := pq.ToBinaryTree()

	tree := NewHuffmanTree(root)

	header := tree.ToHeader()

	if err := os.WriteFile(*output, header, 0644); err != nil {
		log.Fatalf("error writing header to file: %v", err)
	}
}
