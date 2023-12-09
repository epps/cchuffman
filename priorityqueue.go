package main

import (
	"fmt"
	"os"
)

func NewPriorityQueue(list []*FrequencyNode) *PriorityQueue {
	pq := &PriorityQueue{
		nodes: make([]*FrequencyNode, 0),
	}

	for _, node := range list {
		pq.Insert(node)
	}

	return pq
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
	pq.nodes = pq.nodes[:len(pq.nodes)-1]

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

	if parentNode.freq > node.freq || (parentNode.freq == node.freq && parentNode.char > node.char) {
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

	// Cover the edge case in which the left node is present but the right node index is out of bounds
	if rightIdx >= len(pq.nodes) {
		// The left node's frequency is either:
		// less than the current node's frequency
		// is equal to the current node's frequency and its char is less than the current's node's char
		if leftNode.freq < node.freq || (leftNode.freq == node.freq && node.char > leftNode.char) {
			pq.nodes[idx] = leftNode
			pq.nodes[leftIdx] = node
			pq.Down(leftIdx)
		}
		return
	}

	rightNode := pq.nodes[rightIdx]

	// The right node is the smaller than the left node if:
	// * The right node's frequency is less than the left node's frequency
	// * The right and left node's frequencies are equal and the right node's char is less than the left node's char
	// If the right node is smaller, then we swap if:
	// * The current node's frequency is greater than the right node's frequency
	// * The current node and right node's frequencies are equal and the current node's char is greater than the right node's char
	isRightFreqLT := leftNode.freq > rightNode.freq
	isLeftRightFreqEq := leftNode.freq == rightNode.freq
	isRightCharLT := leftNode.char > rightNode.char
	isNodeGTRight := node.freq > rightNode.freq
	isNodeEqRight := node.freq == rightNode.freq
	isNodeCharGTRight := node.char > rightNode.char
	isNodeGTLeft := node.freq > leftNode.freq
	isNodeEqLeft := node.freq == leftNode.freq
	isNodeCharGTLeft := node.char > leftNode.char

	if (isRightFreqLT || (isLeftRightFreqEq && isRightCharLT)) && (isNodeGTRight || (isNodeEqRight && isNodeCharGTRight)) {
		pq.nodes[idx] = rightNode
		pq.nodes[rightIdx] = node
		pq.Down(rightIdx)
	} else if (!isRightFreqLT || (isLeftRightFreqEq && !isRightCharLT)) && (isNodeGTLeft || (isNodeEqLeft && isNodeCharGTLeft)) {
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

func (pq *PriorityQueue) Log(filename string) error {
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

	definitions := ""
	connections := ""

	for idx, node := range pq.nodes {
		definitions += fmt.Sprintf(` node_%d[label="char: %q\nrune: %d\nfreq: %d"];`, idx, node.char, node.char, node.freq)
		leftIdx := pq.leftIdx(idx)
		rightIdx := pq.rightIdx(idx)
		if leftIdx < len(pq.nodes) {
			connections += fmt.Sprintf(` node_%d -- node_%d;`, idx, leftIdx)
		}
		if rightIdx < len(pq.nodes) {
			connections += fmt.Sprintf(` node_%d -- node_%d;`, idx, rightIdx)
		}
	}

	f.WriteString(fmt.Sprintf(`graph {
		%s
		%s
	}`, definitions, connections))

	f.Sync()
	return nil
}

func (pq *PriorityQueue) ToBinaryTree() *FrequencyNode {
	var root *FrequencyNode

	for {
		if len(pq.nodes) == 1 {
			root = pq.nodes[0]
			break
		}

		a := pq.Pop()
		b := pq.Pop()
		c := &FrequencyNode{
			freq:  a.freq + b.freq,
			left:  a,
			right: b,
		}

		pq.Insert(c)
	}

	return root
}
