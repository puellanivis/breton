// Package rand provides a wrapper around math/rand that uses crypto/rand for generating seeds.
package rand

import (
	crand "crypto/rand"
	"time"
)

var globalRand = New(NewSource(genSeed()))

func genSeed() int64 {
	// it doesn't matter that this will always read in little-endian mode.

	b := make([]byte, 8)
	if _, err := crand.Read(b); err != nil {
		panic(err)
	}

	u := int64(b[0] &^ 0x80) // clear sign bit to ensure no overflows
	for _, b := range b[1:] {
		u <<= 8
		u |= int64(b)
	}

	return u
}

func SecureSeed() {
	globalRand.SecureSeed()
}

func Reseeder(d time.Duration) {
	globalRand.Reseeder(d)
}
