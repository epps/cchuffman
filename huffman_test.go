package main

import (
	"testing"
	"unicode/utf8"
)

func TestFrequencyTable(t *testing.T) {
	type testCase struct {
		char         string
		expectedFreq int
	}

	tests := []testCase{
		{
			char:         "a",
			expectedFreq: 10,
		},
		{
			char:         "b",
			expectedFreq: 9,
		},
		{
			char:         "c",
			expectedFreq: 8,
		},
		{
			char:         "ę",
			expectedFreq: 6,
		},
		{
			char:         "i",
			expectedFreq: 2,
		},
		{
			char:         "Ü",
			expectedFreq: 6,
		},
		{
			char:         "Š",
			expectedFreq: 4,
		},
		{
			char:         "Z",
			expectedFreq: 11,
		},
		{
			char:         "k",
			expectedFreq: 0,
		},
		{
			char:         "m",
			expectedFreq: 0,
		},
	}

	ft := NewFrequencyTable("frequency-test.txt")

	if err := ft.Populate(); err != nil {
		t.Fatalf("failed to populate table: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.char, func(t *testing.T) {
			r, _ := utf8.DecodeRuneInString(tt.char)
			actualCount := ft.Get(r)

			if actualCount != tt.expectedFreq {
				t.Fatalf("expected count for %s to be %d, but received %d", tt.char, tt.expectedFreq, actualCount)
			}
		})
	}
}

func TestPriorityQueue(t *testing.T) {
	ft := NewFrequencyTable("frequency-test.txt")

	if err := ft.Populate(); err != nil {
		t.Fatalf("failed to populate table: %v", err)
	}

	nodes := ft.ToList()

	pq := NewPriorityQueue(nodes)

	previousNode := pq.Pop()
	counter := 0
	var currentNode *FrequencyNode
	for {
		currentNode = pq.Pop()
		counter += 1

		if currentNode == nil {
			break
		}

		if currentNode.freq < previousNode.freq || (currentNode.freq == previousNode.freq && currentNode.char < previousNode.char) {
			t.Fatalf("expected %q to have a higher frequency than %q", currentNode.char, previousNode.char)
		}

		previousNode = currentNode
	}
}

func TestBinaryTree(t *testing.T) {
	// See https://opendsa-server.cs.vt.edu/ODSA/Books/CS3/html/Huffman.html#freqexamp
	// for test tree walkthrough and structure
	nodes := []*FrequencyNode{
		{
			char: 67, // C
			freq: 32,
		},
		{
			char: 68, // D
			freq: 42,
		},
		{
			char: 69, // E
			freq: 120,
		},
		{
			char: 75, // K
			freq: 7,
		},
		{
			char: 76, // L
			freq: 42,
		},
		{
			char: 77, // M
			freq: 24,
		},
		{
			char: 85, // U
			freq: 37,
		},
		{
			char: 90, // Z
			freq: 2,
		},
	}

	pq := NewPriorityQueue(nodes)

	tree := NewHuffmanTree(pq.ToBinaryTree())

	inOrderNodes := tree.TraverseInOrder()

	first := *inOrderNodes[0]
	second := *inOrderNodes[1]
	last := *inOrderNodes[len(inOrderNodes)-1]

	if first.char != 69 && first.freq != 120 && first.IsLeaf() {
		t.Fatalf("Expected leaf node with char 69 ('E') with frequency 120, but received char %q with frequency %d", first.char, first.freq)
	}

	if second.char != 0 && second.freq != 306 && !second.IsLeaf() {
		t.Fatalf("Expected internal node with char 0 with frequency 306, but received char %q with frequency %d", second.char, second.freq)
	}

	if last.char != 77 && last.freq != 24 && last.IsLeaf() {
		t.Fatalf("Expected leaf node char 77 ('M') with frequency 24, but received char %q with frequency %d", last.char, last.freq)
	}
}
