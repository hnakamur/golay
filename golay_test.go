package golay

import (
	"bytes"
	"sync"
	"testing"
)

func TestEncodeDecode(t *testing.T) {
	const n = 4
	const max = 0xffffff
	testedCount := make([]int, n)
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()

			input := make([]byte, 3)
			encoded := make([]byte, 0, 6)
			decoded := make([]byte, 0, 3)

			for j := i; j <= max; j += n {
				input[0] = byte(j >> 16)
				input[1] = byte((j >> 8) & 0xff)
				input[2] = byte(j & 0xff)
				encoded2 := Encode(input, encoded)
				errCnt, got := Decode(encoded2, decoded)
				if errCnt != 0 {
					t.Errorf("unexpected errCnt=%d, got=0x%06x, want=0x%06x, encoded=0x%012x",
						errCnt, got, input, encoded)
				}
				if want := input; !bytes.Equal(got, want) {
					t.Errorf("unexpected decode result, got=0x%06x, want=0x%06x, encoded=0x%012x",
						got, input, encoded)
				}
				testedCount[i]++
			}
		}(i)
	}
	wg.Wait()

	var total int
	for _, c := range testedCount {
		total += c
	}
	if got, want := total, max+1; got != want {
		t.Errorf("test count unmatch, got=%d, want=%d", got, want)
	}
}
