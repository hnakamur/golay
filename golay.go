// Package golay provides golay 23/12 error correction encoding and decoding
package golay

// Encode encodes n bytes of data into 2n coded bytes.
// The output is appended to out and returns the result
// of append the output to out.
// n = len(in) must be a multiple 3.
func Encode(in, out []byte) []byte {
	if len(in)%3 != 0 {
		panic("len(in) must be a multiple of 3")
	}
	for len(in) >= 3 {
		out = encode24(in, out)
		in = in[3:]
	}
	return out
}

// encode 3 bytes data into 6 bytes of coded data
func encode24(in, out []byte) []byte {
	v := uint16(in[0]) | uint16(in[1]&0x0F)<<8
	syn := golay23EncodeTable[v]
	out = append(out, uint8(syn&0xFF))
	out = append(out, (in[0]&0x1F)<<3|byte(syn>>8))
	out = append(out, (in[0]&0xE0)>>5|(in[1]&0x0F)<<3)

	v = uint16(in[2]) | uint16(in[1]&0xF0)<<4
	syn = golay23EncodeTable[v]
	out = append(out, uint8(syn&0xFF))
	out = append(out, (in[2]&0x1F)<<3|byte(syn>>8))
	out = append(out, (in[2]&0xE0)>>5|(in[1]&0xF0)>>1)
	return out
}

// Decode decodes n bytes of coded data into n/2 bytes of original data.
// The output is appended to out and returns the result
// of append the output to out.
// Also the number of 12 bit words that required correction
// is returned.
// n = len(in) must be a multiple of 6.
func Decode(in, out []byte) (int, []byte) {
	if len(in)%6 != 0 {
		panic("len(in) must be a multiple of 6")
	}
	var errcount, c int
	for len(in) >= 6 {
		c, out = decode24(in, out)
		errcount += c
		in = in[6:]
	}
	return errcount, out
}

// decode 6 bytes of coded data into 3 bytes of original data
// returns the number of words corrected (0, 1 or 2) and maybe modified out.
func decode24(in, out []byte) (int, []byte) {
	var errcount int

	v := uint16(in[2]&0x7F)<<5 | uint16(in[1]&0xF8)>>3
	syn := golay23EncodeTable[v]
	syn ^= uint16(in[0]) | uint16(in[1]&0x07)<<8
	e := golay23DecodeTable[syn]
	if e != 0 {
		errcount++
		v ^= e
	}
	out = append(out, byte(v&0xFF))
	out1 := byte(v >> 8)

	v = uint16(in[5]&0x7F)<<5 | uint16(in[4]&0xF8)>>3
	syn = golay23EncodeTable[v]
	syn ^= uint16(in[3]) | uint16(in[4]&0x07)<<8
	e = golay23DecodeTable[syn]
	if e != 0 {
		errcount++
		v ^= e
	}
	out = append(out, out1|(byte(v>>4)&0xF0))
	out = append(out, byte(v&0xFF))

	return errcount, out
}
