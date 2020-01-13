package golay

import (
	"bytes"
	crand "crypto/rand"
	"encoding/binary"
	"flag"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"
)

var seed = flag.Int64("seed", 0, "random seed")

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

func TestEncodeDecodeNoError(t *testing.T) {
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
	t.Logf("tested %d patterns", total)
}

func TestEncodeDecodeOneError(t *testing.T) {
	if *seed == 0 {
		*seed = newRandSeed()
	}
	t.Logf("seed=%d", *seed)
	rnd := rand.New(rand.NewSource(*seed))

	encodeInput := make([]byte, 3)
	encoded := make([]byte, 0, 6)
	decodeInput := make([]byte, 6)
	decoded := make([]byte, 0, 3)

	const n = 10
	const m = 3
	var unexpected int
	for i := 0; i < n; i++ {
		if _, err := rnd.Read(encodeInput); err != nil {
			t.Fatal(err)
		}

		encoded2 := Encode(encodeInput, encoded[:0])
		for j := 0; j < m; j++ {
			copy(decodeInput, encoded2)
			k := rnd.Intn(len(decodeInput))
			for {
				b := byte(rnd.Intn(256))
				if b != decodeInput[k] {
					decodeInput[k] = b
					break
				}
			}

			gotErrCnt, gotDecoded := Decode(decodeInput, decoded[:0])
			wantErrCnt := 1
			wantDecoded := encodeInput
			if gotErrCnt != wantErrCnt || !bytes.Equal(gotDecoded, wantDecoded) {
				t.Errorf("unexpected errCnt or decode result, input=%06x, encoded=%012x, errInjected=%012x, gotErrCnt=%d, wantErrCnt=%d, gotDecoded=%06x, wantDecoded=%06x",
					encodeInput, encoded2, decodeInput, gotErrCnt, wantErrCnt, gotDecoded, wantDecoded)
				unexpected++
			}
		}
	}
	if unexpected > 0 {
		t.Logf("unexpected case count=%d, total=%d", unexpected, n*m)
	}
}

func newRandSeed() int64 {
	var b [8]byte
	if _, err := crand.Read(b[:]); err != nil {
		return time.Now().UnixNano()
	}
	return int64(binary.BigEndian.Uint64(b[:]))
}
