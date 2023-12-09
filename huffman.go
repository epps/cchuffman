package main

import (
	"bytes"
	"fmt"
	"os"
)

const CONTROL_CHAR rune = '⁂'

type FrequencyNode struct {
	char  rune
	freq  int
	left  *FrequencyNode
	right *FrequencyNode
}

func (fn *FrequencyNode) IsLeaf() bool {
	return fn.left == nil && fn.right == nil
}

func NewHuffmanTree(root *FrequencyNode) *HuffmanTree {
	return &HuffmanTree{
		root: root,
	}
}

type HuffmanTree struct {
	root *FrequencyNode
}

func (hf *HuffmanTree) Log(filename string) error {
	if _, err := os.Stat("graphviz"); os.IsNotExist(err) {
		if err := os.Mkdir("graphviz", 0755); err != nil {
			return err
		}
	}
	f, err := os.Create(fmt.Sprintf("graphviz/%s.dot", filename))
	if err != nil {
		return err
	}
	defer f.Close()

	queue := []*FrequencyNode{hf.root}

	definitions := ""
	connections := ""
	for {
		if len(queue) == 0 {
			break
		}

		node := queue[0]
		queue = queue[1:]

		if node.IsLeaf() {
			definitions += fmt.Sprintf(` node_%d_%d[label="char: %q\nrune: %d\nfreq: %d"];`, node.freq, node.char, node.char, node.char, node.freq)
		} else {
			definitions += fmt.Sprintf(` node_%d_%d[label="weight %d"];`, node.freq, node.char, node.freq)
			if node.left != nil {
				queue = append(queue, node.left)
				left := *node.left
				connections += fmt.Sprintf(` node_%d_%d -- node_%d_%d[label="%d"];`, node.freq, node.char, left.freq, left.char, 0)
			}

			if node.right != nil {
				queue = append(queue, node.right)
				right := *node.right
				connections += fmt.Sprintf(` node_%d_%d -- node_%d_%d[label="%d"];`, node.freq, node.char, right.freq, right.char, 1)
			}
		}
	}

	f.WriteString(fmt.Sprintf(`graph {
		%s
		%s
	}`, definitions, connections))

	f.Sync()

	return nil
}

func (hf *HuffmanTree) ToLookupTable() map[rune]string {
	table := make(map[rune]string)

	var traverse func(n *FrequencyNode, code string)
	traverse = func(n *FrequencyNode, code string) {
		if n == nil {
			return
		}

		// In-order traversal
		traverse(n.left, code+"0")
		table[n.char] = code
		traverse(n.right, code+"1")
	}

	traverse(hf.root, "")

	return table
}

func (hf *HuffmanTree) ToHeader() []byte {
	header := new(bytes.Buffer)
	bitWriter := NewBitWriter(header)

	var traverse func(n *FrequencyNode)
	traverse = func(n *FrequencyNode) {
		if n == nil {
			return
		}

		// Pre-order traversal
		if n.IsLeaf() {
			bitWriter.WriteBit(One)
			bitWriter.WriteRune(n.char)
		} else {
			bitWriter.WriteBit(Zero)
		}
		traverse(n.left)
		traverse(n.right)
	}

	traverse(hf.root)

	bitWriter.Flush(One)

	bitWriter.WriteRune(CONTROL_CHAR)

	return header.Bytes()
}
