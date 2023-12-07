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

	inOrderNodes := make([]*FrequencyNode, 0)

	// Perform in-order depth-first traversal to capture a representation
	// of the known Huffman tree to verify the tree's structure is correct.
	// The true litmus test of success will be the prefix-free table; however,
	// given the breakdown of the steps (i.e. generate the binary tree _before
	// generating the lookup table), it's helpful to verify the tree's correctness
	// independent of the Huffman codes.
	var traverse func(n *FrequencyNode)
	traverse = func(n *FrequencyNode) {
		if n == nil {
			return
		}

		traverse(n.left)
		inOrderNodes = append(inOrderNodes, n)
		traverse(n.right)
	}

	traverse(tree.root)

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

func TestLookupTable(t *testing.T) {
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

	expectedCodes := []struct {
		char rune
		code string
	}{
		{
			char: 67, // C
			code: "1110",
		},
		{
			char: 68, // D
			code: "101",
		},
		{
			char: 69, // E
			code: "0",
		},
		{
			char: 75, // K
			code: "111101",
		},
		{
			char: 76, // L
			code: "110",
		},
		{
			char: 77, // M
			code: "11111",
		},
		{
			char: 85, // U
			code: "100",
		},
		{
			char: 90, // Z
			code: "111100",
		},
	}

	pq := NewPriorityQueue(nodes)

	tree := NewHuffmanTree(pq.ToBinaryTree())

	table := tree.ToLookupTable()

	for _, expected := range expectedCodes {
		actualCode := table[expected.char]

		if expected.code != actualCode {
			t.Fatalf("Expected %d char to have code %s, but received code %s", expected.char, expected.code, actualCode)
		}
	}

}

func TestToHeader(t *testing.T) {
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

	header := tree.ToHeader()

	expected := "01E001U1D01L01C001Z1K1M⁂"

	if string(header) != expected {
		t.Fatalf("Expected header to be %s, but received %s", expected, header)
	}
}
