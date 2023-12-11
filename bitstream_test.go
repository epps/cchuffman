package main

import (
	"bytes"
	"io"
	"testing"
)

func TestWritingAndReading(t *testing.T) {
	testHeader := []rune{'0', '1', 'E', '0', '0', '1', 'U', '1', 'D', '0', '1', 'L', '0', '1', 'C', '0', '0', '1', 'Z', '1', 'K', '1', 'M', '⁂'}

	buf := bytes.Buffer{}

	writer := NewBitWriter(&buf)

	for _, r := range testHeader {
		switch r {
		case '0':
			if err := writer.WriteBit(Zero); err != nil {
				t.Error(err)
			}
		case '1':
			if err := writer.WriteBit(One); err != nil {
				t.Error(err)
			}
		default:
			if err := writer.WriteRune(r); err != nil {
				t.Error(err)
			}
		}
	}

	writer.Flush(One)

	reader := NewBitReader(&buf)

	for _, expected := range testHeader {
		switch expected {
		case '0':
			bit, err := reader.ReadBit()
			if err != nil {
				t.Error(err)
			}
			if bit {
				t.Errorf("Expected zero but received %v", bit)
			}
		case '1':
			bit, err := reader.ReadBit()
			if err != nil {
				t.Error(err)
			}
			if !bit {
				t.Errorf("Expected one but received %v", bit)
			}
		default:
			actual, err := reader.ReadRune()
			if err != nil && err != io.EOF {
				t.Error(err)
			}
			if expected != actual {
				t.Errorf("Expected %q but received %q", expected, actual)
			}
		}
	}

	if err := reader.Flush(); err != nil && err != io.EOF {
		t.Error(err)
	}
}
