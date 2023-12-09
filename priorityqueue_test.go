package main

import "testing"

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
