package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"
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

type FrequencyNode struct {
	char  rune
	freq  int
	left  *FrequencyNode
	right *FrequencyNode
}

func (ft *FrequencyTable) ToList() []*FrequencyNode {
	list := make([]*FrequencyNode, 0)

	for char, freq := range ft.table {
		list = append(list, &FrequencyNode{char: char, freq: freq})
	}

	return list
}

func NewPriorityQueue(list []*FrequencyNode) *PriorityQueue {
	sort.Slice(list, func(i, j int) bool {
		return list[i].freq < list[j].freq
	})

	return &PriorityQueue{
		nodes: list,
	}
}

type PriorityQueue struct {
	nodes []*FrequencyNode
}

func (pq *PriorityQueue) Insert(val *FrequencyNode) {
	pq.nodes = append(pq.nodes, val)
	pq.Up(len(pq.nodes) - 1)
}

func (pq *PriorityQueue) Pop() *FrequencyNode {
	if len(pq.nodes) == 0 {
		return nil
	}

	out := pq.nodes[0]

	if len(pq.nodes) == 1 {
		pq.nodes = []*FrequencyNode{}
		return out
	}

	last := pq.nodes[len(pq.nodes)-1]

	pq.nodes[0] = last
	pq.nodes = pq.nodes[:len(pq.nodes)]

	pq.Down(0)

	return out
}

func (pq *PriorityQueue) Up(idx int) {
	if idx == 0 {
		return
	}

	parentIdx := pq.parentIdx(idx)
	parentNode := pq.nodes[parentIdx]
	node := pq.nodes[idx]

	if parentNode.freq > node.freq {
		pq.nodes[idx] = parentNode
		pq.nodes[parentIdx] = node
		pq.Up(parentIdx)
	}
}

func (pq *PriorityQueue) Down(idx int) {
	if idx >= len(pq.nodes) {
		return
	}

	leftIdx := pq.leftIdx(idx)

	// Given Heaps are complete from left to right, if the left child's index
	// is at or beyond the length of the heap, then we cannot continue.
	if leftIdx >= len(pq.nodes) {
		return
	}

	rightIdx := pq.rightIdx(idx)

	node := pq.nodes[idx]
	leftNode := pq.nodes[leftIdx]
	rightNode := pq.nodes[rightIdx]

	if leftNode.freq > rightNode.freq && node.freq > rightNode.freq {
		pq.nodes[idx] = rightNode
		pq.nodes[rightIdx] = node
		pq.Down(rightIdx)
	} else if rightNode.freq > leftNode.freq && node.freq > leftNode.freq {
		pq.nodes[idx] = leftNode
		pq.nodes[leftIdx] = node
		pq.Down(leftIdx)
	}
}

func (pq *PriorityQueue) parentIdx(idx int) int {
	return (idx - 1) / 2
}

func (pq *PriorityQueue) leftIdx(idx int) int {
	return (2 * idx) + 1
}

func (pq *PriorityQueue) rightIdx(idx int) int {
	return (2 * idx) + 2
}
