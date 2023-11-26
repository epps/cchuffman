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
