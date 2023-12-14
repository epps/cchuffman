package main

import (
	"bytes"
	"crypto/md5"
	"io"
	"os"
	"reflect"
	"testing"
)

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
		t.Errorf("Expected leaf node with char 69 ('E') with frequency 120, but received char %q with frequency %d", first.char, first.freq)
	}

	if second.char != 0 && second.freq != 306 && !second.IsLeaf() {
		t.Errorf("Expected internal node with char 0 with frequency 306, but received char %q with frequency %d", second.char, second.freq)
	}

	if last.char != 77 && last.freq != 24 && last.IsLeaf() {
		t.Errorf("Expected leaf node char 77 ('M') with frequency 24, but received char %q with frequency %d", last.char, last.freq)
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
			t.Errorf("Expected %d char to have code %s, but received code %s", expected.char, expected.code, actualCode)
		}
	}

}

func TestHeader(t *testing.T) {
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

	inputTree := NewHuffmanTree(pq.ToBinaryTree())

	header := bytes.Buffer{}

	bitWriter := NewBitWriter(&header)

	inputTree.WriteHeader(bitWriter)

	// The string representation of the pre-order traversal of the tree should
	// look like this: "01E001U1D01L01C001Z1K1M⁂"
	// "01 E 001 U 1 D 01 L01 C 001 Z 1 K 1 M ⁂"
	// This includes
	// * 8 1-byte runes (E, U, D, L, C, Z, K, M) which accounts for 8 bytes
	// * 1 3-byte rune (⁂) which accounts for 3 bytes
	// * 15 tree-traversal bits (0, 1, 0, 0, 1, 1, 0, 1, 0, 1, 0, 0, 1, 1, 1), which are padded
	// 1 final bit (1) to account for 2 bytes
	// This makes a total of 13 bytes

	if header.Len() != 13 {
		t.Errorf("Expected header to be 13 bytes long, but received %d bytes", header.Len())
	}

	f, err := os.Create("input.txt")
	if err != nil {
		t.Error(err)
	}
	f.Write(header.Bytes())
	f.Close()

	inputFile, err := os.Open("input.txt")
	if err != nil {
		t.Error(err)
	}
	defer inputFile.Close()

	bitReader := NewBitReader(inputFile)

	outputTree := &HuffmanTree{}

	outputTree.ReadHeader(bitReader)
	if err != nil {
		t.Error(err)
	}

	r, err := bitReader.ReadRune()
	if err != nil {
		t.Error(err)
	}

	err = bitReader.Flush()
	if err != nil && err != io.EOF {
		t.Error(err)
	}

	if !reflect.DeepEqual(inputTree.ToLookupTable(), outputTree.ToLookupTable()) {
		t.Error("expected input lookup table to equal output lookup table", inputTree.ToLookupTable(), outputTree.ToLookupTable())
	}

	if r != rune('⁂') {
		t.Errorf("expected header control character (⁂) but received %q instead", r)
	}

	if err != io.EOF {
		t.Errorf("expected EOF error but received %v instead", err)
	}
}

func TestCompression(t *testing.T) {
	input := "les-mis-test.txt"
	output := "output.txt"
	decompressed := "original.txt"
	encoder := NewHuffmanEncoder(input, output)

	if err := encoder.Encode(); err != nil {
		t.Errorf("failed to compress %s: %v", input, err)
	}

	decoder := NewHuffmanDecoder(output, decompressed)

	if err := decoder.Decode(); err != nil {
		t.Errorf("failed to decompress %s: %v", output, err)
	}

	inputHash := md5.New()
	inputFile, err := os.Open(input)
	if err != nil {
		t.Error(err)
	}
	defer inputFile.Close()
	if _, err := io.Copy(inputHash, inputFile); err != nil {
		t.Error(err)
	}

	decompressedHash := md5.New()
	decompressedFile, err := os.Open(decompressed)
	if err != nil {
		t.Error(err)
	}
	defer decompressedFile.Close()
	if _, err := io.Copy(decompressedHash, decompressedFile); err != nil {
		t.Error(err)
	}

	if string(inputHash.Sum(nil)) != string(decompressedHash.Sum(nil)) {
		t.Error("Expected decompressed to be identical to original file")
	}
}
