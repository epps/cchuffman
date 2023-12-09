package main

import (
	"fmt"
	"io"
	"unicode/utf8"
)

// *NOTE* It was clear that to accomplish the encoding, the program would need to:
// * write single bits to a buffer/stream/file to signify the tree traversal
// * write runes to a buffer/stream/file to identify the encoded characters for the decoder
// * write prefix-free codes bit-by-bit to a buffer/stream/file to compress the input
// Given my gross ignorance of hands-on bit manipulation, I searched for a Go package
// that would teach me what these items might look like in code. I stumbled upon
// [go-bitstream](https://github.com/dgryski/go-bitstream), which was both small and
// easy for me to read and digest. The below code is essentially a plargiarism of go-bitstream
// for the purpose of learning and completing this coding challenge

type Bit bool

const (
	Zero Bit = false
	One  Bit = true
)

// BitWriter wraps io.Writer with methods to write single bits and/or runes
type BitWriter struct {
	writer io.Writer
	// alignment keeps track of the write position of the current buffer and
	// begins with MSB (most significant bit)
	alignment uint8
	buffer    [1]byte
}

func NewBitWriter(writer io.Writer) *BitWriter {
	return &BitWriter{
		writer:    writer,
		alignment: 8,
		buffer:    [1]byte{},
	}
}

func (bw *BitWriter) WriteBit(bit Bit) error {
	// add bit to buffer by setting the LSB and shifting buffer left

	// ex. buffer is currently 1000 0000 (i.e. 1 bit has been written) and bit is 1
	// 1 in binary is 0000 0001
	// alignment is 7 after 1 write, so the next write position is 6 (i.e. 1*00 0000)
	// By shifting 1 over 6 times, it becomes 0100 0000
	// With left-shift 1 ready, the |= (bitwise OR compound assignment operator) sets
	// the next bit in the buffer
	if bit {
		bw.buffer[0] |= 1 << (bw.alignment - 1)
	}

	bw.alignment -= 1

	if bw.alignment == 0 {
		if n, err := bw.writer.Write(bw.buffer[:]); n != 1 || err != nil {
			return err
		}
		bw.buffer[0] = 0
		bw.alignment = 8
	}

	return nil
}

func (bw *BitWriter) WriteRune(char rune) error {
	lenBytes := utf8.RuneLen(char)
	if lenBytes == -1 {
		return fmt.Errorf("character %q is not valid rune", char)
	}
	bytes := make([]byte, lenBytes)
	utf8.EncodeRune(bytes, char)

	for _, b := range bytes {
		// right shift the byte by the difference between a complete byte
		// and the current alignment to fill the available bits in the buffer
		// with the most significant bits
		bw.buffer[0] |= b >> (8 - bw.alignment)

		// write the byte to the buffer
		if n, err := bw.writer.Write(bw.buffer[:]); n != 1 || err != nil {
			return err
		}

		// left shift the byte by the current alignment to write the least
		// significant bits were excluded by the right shift operation to
		// the buffer
		bw.buffer[0] = b << bw.alignment
	}

	return nil
}

func (bw *BitWriter) Flush(bit Bit) error {
	for bw.alignment != 8 {
		if err := bw.WriteBit(bit); err != nil {
			return err
		}
	}
	return nil
}
