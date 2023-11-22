package main

import (
	"testing"
	"unicode/utf8"
)

func TestFrequencyTable(t *testing.T) {
	type testCase struct {
		char          string
		expectedCount int
	}

	tests := []testCase{
		{
			char:          "a",
			expectedCount: 10,
		},
		{
			char:          "b",
			expectedCount: 9,
		},
		{
			char:          "c",
			expectedCount: 8,
		},
		{
			char:          "ę",
			expectedCount: 6,
		},
		{
			char:          "i",
			expectedCount: 2,
		},
		{
			char:          "Ü",
			expectedCount: 6,
		},
		{
			char:          "Š",
			expectedCount: 4,
		},
		{
			char:          "Z",
			expectedCount: 11,
		},
		{
			char:          "k",
			expectedCount: 0,
		},
		{
			char:          "m",
			expectedCount: 0,
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

			if actualCount != tt.expectedCount {
				t.Fatalf("expected count for %s to be %d, but received %d", tt.char, tt.expectedCount, actualCount)
			}
		})
	}
}
