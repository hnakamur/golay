package golay

import (
	"bytes"
	"flag"
	"os"
	"sync"
	"testing"

	"golang.org/x/exp/rand"

	randutil "github.com/hnakamur/randutil/v3"
)

var seed = flag.Uint64("seed", 0, "random seed")

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

func TestEncodeDecode3bitsError(t *testing.T) {
	if *seed == 0 {
		*seed = randutil.NewSeed()
	}
	t.Logf("seed=%d", *seed)
	src := rand.NewSource(*seed)
	rnd := rand.New(src)

	encodeInput := make([]byte, 3)
	encoded := make([]byte, 0, 6)
	decodeInput := make([]byte, 6)
	decoded := make([]byte, 0, 3)

	const n = 100
	const m = 10
	var unexpectedErrCnt, unexpectedDecoded, unexpected int
	for i := 0; i < n; i++ {
		if _, err := rnd.Read(encodeInput); err != nil {
			t.Fatal(err)
		}

		encoded2 := Encode(encodeInput, encoded[:0])
		for j := 0; j < m; j++ {
			copy(decodeInput, encoded2)
			positions := randutil.MultiIntnNoDup(src, 3, 6*8)
			for _, pos := range positions {
				i := pos / 8
				shift := pos % 8
				decodeInput[i] ^= 1 << shift
			}

			gotErrCnt, gotDecoded := Decode(decodeInput, decoded[:0])
			wantErrCnt := 0
			wantDecoded := encodeInput

			unmatchErrCnt := gotErrCnt == 0
			unmatchDecoded := !bytes.Equal(gotDecoded, wantDecoded)
			var unmatchText string
			if unmatchErrCnt && unmatchDecoded {
				unmatchText = "errCnt and decode result"
			} else if unmatchErrCnt {
				unmatchText = "errCnt"
			} else if unmatchDecoded {
				unmatchText = "decode result"
			}

			if unmatchErrCnt {
				unexpectedErrCnt++
			}
			if unmatchDecoded {
				unexpectedDecoded++
			}
			if unmatchErrCnt {
				t.Logf("unexpected %s, input=%06x, encoded=%012x, errInjected=%012x, gotErrCnt=%d, wantErrCnt=%d, gotDecoded=%06x, wantDecoded=%06x",
					unmatchText, encodeInput, encoded2, decodeInput, gotErrCnt, wantErrCnt, gotDecoded, wantDecoded)
				unexpected++
			} else if unmatchDecoded {
				t.Errorf("unexpected %s, input=%06x, encoded=%012x, errInjected=%012x, gotErrCnt=%d, wantErrCnt=%d, gotDecoded=%06x, wantDecoded=%06x",
					unmatchText, encodeInput, encoded2, decodeInput, gotErrCnt, wantErrCnt, gotDecoded, wantDecoded)
				unexpected++
			}
		}
	}
	if unexpected > 0 {
		t.Logf("unexpected=%d, unexpectedErrCnt=%d, unexpectedDecoded=%d, total=%d",
			unexpected, unexpectedErrCnt, unexpectedDecoded, n*m)
	}
}
