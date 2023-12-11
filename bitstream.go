package main

import (
	"bytes"
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
// easy for me to read and digest. The below code is essentially a plagiarism of go-bitstream
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

type BitReader struct {
	reader    io.Reader
	alignment uint8
	buffer    [1]byte
}

func NewBitReader(r io.Reader) *BitReader {
	return &BitReader{
		reader:    r,
		alignment: 0,
		buffer:    [1]byte{},
	}
}

func (br *BitReader) ReadBit() (Bit, error) {
	// if buffer is complete, push the byte to the reader and reset
	if br.alignment == 0 {
		if n, err := br.reader.Read(br.buffer[:]); n != 1 || err != nil {
			return Zero, err
		}
		br.alignment = 8
	}
	br.alignment -= 1
	// bitwise AND extracts the most significant bit from the buffer because
	// 0x80 in binary is 1000 0000, which means the result will either be
	// 0x80 in the case the buffer's MSB is also 1 or 0 in the case it's not
	bit := (br.buffer[0] & 0x80)
	// left shift compound assignment operator left shift the buffer by one
	// after the read and assigns the result back to the buffer for the next
	// read operation
	br.buffer[0] <<= 1
	return bit != 0, nil
}

func (br *BitReader) ReadRune() (rune, error) {
	// runes can be multiple bytes, so keeping a buffer external to
	// the reader's buffer, which is just 1 byte, is necessary to handle
	// multi-byte runes (e.g. the header's control character '⁂', which is
	// 3 bytes)
	rBuff := bytes.Buffer{}
	for {
		// If the buffer exceeds a logical maximum, then something has errored
		// and we should break to avoid an infinite loop.
		// TODO: understand the difference/implications of using this over
		// utf8.MaxRune
		if rBuff.Len() > utf8.UTFMax {
			return 0, fmt.Errorf("failed to read complete rune; rune exceeds max UTF")
		}
		// Given the logic of ReadRune should only ever operate at the level of 1 byte
		// per iteration, either DecodeRune will return a value rune after 1 or more
		// iterations or it should error
		r, size := utf8.DecodeRune(rBuff.Bytes())
		if r != utf8.RuneError && size != 0 {
			return r, nil
		}
		// Alignment is zero with the reader's buffer has a complete byte to operate
		// on; reading 1 or more complete bytes will not change the alignment so no
		// resetting/decrementing happens
		if br.alignment == 0 {
			n, err := br.reader.Read(br.buffer[:])
			if n != 1 || (err != nil && err != io.EOF) {
				br.buffer[0] = 0
				return rune(br.buffer[0]), err
			}
			// TODO: handle this case
			if err == io.EOF {
				err = nil
			}
			if n, err := rBuff.Write(br.buffer[:]); n != 1 || err != nil {
				return 0, err
			}
		} else {
			// If the alignment is not zero, the buffer contains 1 or more bits that need to be
			// saved before reading another byte into the reader's buffer
			currentBuf := br.buffer[0]

			if n, err := br.reader.Read(br.buffer[:]); n != 1 || (err != nil && err != io.EOF) {
				return 0, err
			}
			// Right shifting the newly filled reader buffer by the alignment and assigning the result
			// to the buffer that saved the partial bits to complete the byte
			currentBuf |= br.buffer[0] >> br.alignment

			// Left shifting by the difference between 8 (i.e. an alignment that points to the MSB) and
			// the current alignment to prepare the reader for the next read operation
			br.buffer[0] <<= (8 - br.alignment)

			if err := rBuff.WriteByte(currentBuf); err != nil {
				return 0, err
			}
		}
	}
}

func (br *BitReader) Flush() error {
	for br.alignment != 8 {
		if _, err := br.ReadBit(); err != nil {
			return err
		}
	}
	return nil
}
