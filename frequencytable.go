package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"text/tabwriter"
	"unicode/utf8"
)

func NewFrequencyTable(filename string) *FrequencyTable {
	return &FrequencyTable{
		filename: filename,
		table:    make(map[rune]int),
	}
}

type FrequencyTable struct {
	filename string
	table    map[rune]int
}

func (ft *FrequencyTable) Populate() error {
	file, err := os.Open(ft.filename)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %v", ft.filename, err)
	}
	defer func(f *os.File) {
		if err := f.Close(); err != nil {
			log.Fatalf("failed to close file %s: %v\n", f.Name(), err)
		}
	}(file)

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanRunes)

	for scanner.Scan() {
		r, _ := utf8.DecodeRuneInString(scanner.Text())
		ft.table[r] += 1
	}

	return nil
}

func (ft *FrequencyTable) Log() error {
	writer := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)

	count := 0
	for r, freq := range ft.table {
		count += 1
		fmt.Fprintf(writer, "%d.\t%q\t%d\n", count, r, freq)
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush tab writer: %v", err)
	}

	return nil
}

func (ft *FrequencyTable) Get(r rune) int {
	return ft.table[r]
}

func (ft *FrequencyTable) ToList() []*FrequencyNode {
	list := make([]*FrequencyNode, 0)

	for char, freq := range ft.table {
		list = append(list, &FrequencyNode{char: char, freq: freq})
	}

	return list
}
