package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"unicode/utf8"
)

const CONTROL_CHAR rune = '⁂'
const BITS_IN_BYTE = 1024

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

		// Pre-order traversal
		table[n.char] = code
		traverse(n.left, code+"0")
		traverse(n.right, code+"1")
	}

	traverse(hf.root, "")

	return table
}

func (hf *HuffmanTree) WriteHeader(w *BitWriter) {
	var traverse func(n *FrequencyNode)
	traverse = func(n *FrequencyNode) {
		if n == nil {
			return
		}

		// Pre-order traversal
		if n.IsLeaf() {
			w.WriteBit(One)
			w.WriteRune(n.char)
		} else {
			w.WriteBit(Zero)
		}
		traverse(n.left)
		traverse(n.right)
	}

	traverse(hf.root)

	w.WriteRune(CONTROL_CHAR)

	w.Flush(One)
}

func (hf *HuffmanTree) ReadHeader(r *BitReader) error {
	var traverse func(n *FrequencyNode) error
	traverse = func(n *FrequencyNode) error {
		if n.char != 0 {
			return nil
		}

		bit, err := r.ReadBit()
		if err != nil {
			return err
		}

		var left *FrequencyNode
		if bit == Zero {
			left = &FrequencyNode{}
		} else {
			r, err := r.ReadRune()
			if err != nil {
				return err
			}
			left = &FrequencyNode{char: r}
		}
		if err := traverse(left); err != nil {
			return err
		}

		bit, err = r.ReadBit()
		if err != nil {
			return err
		}

		var right *FrequencyNode
		if bit == Zero {
			right = &FrequencyNode{}
		} else {
			r, err := r.ReadRune()
			if err != nil {
				return err
			}
			right = &FrequencyNode{char: r}
		}
		if err := traverse(right); err != nil {
			return err
		}

		n.left = left
		n.right = right

		return nil
	}

	if bit, err := r.ReadBit(); bit != Zero || err != nil {
		return fmt.Errorf("expected to read initial zero bit: %v", err)
	}
	hf.root = &FrequencyNode{}
	return traverse(hf.root)
}

type HuffmanEncoder struct {
	input  string
	output string
}

func NewHuffmanEncoder(input, output string) *HuffmanEncoder {
	return &HuffmanEncoder{
		input:  input,
		output: output,
	}
}

func (e *HuffmanEncoder) Encode() error {
	ft := NewFrequencyTable(e.input)

	if err := ft.Populate(); err != nil {
		return fmt.Errorf("error populating frequency table: %v", err)
	}

	pq := NewPriorityQueue(ft.ToList())

	root := pq.ToBinaryTree()

	tree := NewHuffmanTree(root)

	lookupTable := tree.ToLookupTable()

	inputFile, err := os.Open(e.input)
	if err != nil {
		return err
	}
	defer inputFile.Close()

	outputFile, err := os.Create(e.output)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	writer := NewBitWriter(outputFile)

	tree.WriteHeader(writer)

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

	inputSizeMB := inputInfo.Size() / BITS_IN_BYTE
	outputSizeMB := outputInfo.Size() / BITS_IN_BYTE

	log.Printf("Input %s (%d KB) successfully written to %s (%d KB)", inputInfo.Name(), inputSizeMB, outputInfo.Name(), outputSizeMB)

	return nil
}
