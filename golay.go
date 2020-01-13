// Copyright (c) 2012 Andrew Tridgell, All Rights Reserved
// Copyright (c) 2019 Hiroaki Nakamura, All Rights Reserved
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions
// are met:
//
//  o Redistributions of source code must retain the above copyright
//    notice, this list of conditions and the following disclaimer.
//  o Redistributions in binary form must reproduce the above copyright
//    notice, this list of conditions and the following disclaimer in
//    the documentation and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS
// FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE
// COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT,
// INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
// (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
// SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION)
// HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT,
// STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED
// OF THE POSSIBILITY OF SUCH DAMAGE.
//

// Package golay provides golay 23/12 error correction encoding and decoding
package golay

// Encode encodes n bytes of data into 2n coded bytes.
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
// n = len(in) must be a multiple of 6.
// the number of 12 bit words that required correction
// and maybe modified out is returned.
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
