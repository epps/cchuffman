package main

import (
	"flag"
	"fmt"
	"log"
)

func main() {
	file := flag.String("file", "les-mis.txt", "a file to encode")

	flag.Parse()

	ft := NewFrequencyTable(*file)

	if err := ft.Populate(); err != nil {
		log.Fatalf("error populating frequency table: %v\n", err)
	}

	fmt.Printf("Logging character frequencies for %s:\n", *file)
	if err := ft.Log(); err != nil {
		log.Fatalf("error logging frequency table: %v", err)
	}
}
