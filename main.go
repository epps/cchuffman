package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"unicode/utf8"
)

func main() {
	input := flag.String("input", "les-mis.txt", "the input file to encode")
	output := flag.String("output", "compressed.txt", "the output filepath")

	flag.Parse()

	if err := Encode(input, output); err != nil {
		log.Fatalf("failed to compress %s: %v", *input, err)
	}
}

func Encode(input, output *string) error {
	ft := NewFrequencyTable(*input)

	if err := ft.Populate(); err != nil {
		return fmt.Errorf("error populating frequency table: %v", err)
	}

	pq := NewPriorityQueue(ft.ToList())

	root := pq.ToBinaryTree()

	tree := NewHuffmanTree(root)

	lookupTable := tree.ToLookupTable()

	inputFile, err := os.Open(*input)
	if err != nil {
		return err
	}
	defer inputFile.Close()

	outputFile, err := os.Create(*output)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	tree.WriteHeader(outputFile)

	writer := NewBitWriter(outputFile)

	scanner := bufio.NewScanner(inputFile)
	scanner.Split(bufio.ScanRunes)

	for scanner.Scan() {
		r, _ := utf8.DecodeRuneInString(scanner.Text())
		code, hasRune := lookupTable[r]
		if !hasRune {
			return fmt.Errorf("failed to lookup %q", r)
		}
		for _, c := range code {
			switch c {
			case '0':
				if err := writer.WriteBit(Zero); err != nil {
					return fmt.Errorf("failed to write zero bit to output for char %q", r)
				}
			case '1':
				if err := writer.WriteBit(One); err != nil {
					return fmt.Errorf("failed to write one bit to output for char %q", r)
				}
			default:
				return fmt.Errorf("unrecognized code component: %q", c)
			}
		}
	}

	if err := writer.Flush(One); err != nil {
		return fmt.Errorf("failed to flush writer: %v", err)
	}

	inputInfo, err := inputFile.Stat()
	if err != nil {
		return err
	}
	outputInfo, err := outputFile.Stat()
	if err != nil {
		return err
	}

	inputSizeMB := inputInfo.Size() / 1024
	outputSizeMB := outputInfo.Size() / 1024

	log.Printf("Input %s (%d KB) successfully written to %s (%d KB)", inputInfo.Name(), inputSizeMB, outputInfo.Name(), outputSizeMB)

	return nil
}
