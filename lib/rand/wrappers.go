package rand

import (
	"math/rand"
	"time"
)

// Float32 returns Float32 from the global random source.
func Float32() float32 {
	return globalRand.Float32()
}

// Float64 returns Float64 from the global random source.
func Float64() float64 {
	return globalRand.Float64()
}

// ExpFloat64 returns ExpFloat64 from the global random source.
func ExpFloat64() float64 {
	return globalRand.ExpFloat64()
}

// NormFloat64 returns NormFloat64 from the global random source.
func NormFloat64() float64 {
	return globalRand.NormFloat64()
}

// Int returns Int from the global random source.
func Int() int {
	return globalRand.Int()
}

// Intn returns Intn from the global random source.
func Intn(n int) int {
	return globalRand.Intn(n)
}

// Int31 returns Int31 from the global random source.
func Int31() int32 {
	return globalRand.Int31()
}

// Int31n returns Int31n from the global random source.
func Int31n(n int32) int32 {
	return globalRand.Int31n(n)
}

// Int63 returns Int63 from the global random source.
func Int63() int64 {
	return globalRand.Int63()
}

// Int63n returns Int63n from the global random source.
func Int63n(n int64) int64 {
	return globalRand.Int63n(n)
}

// Uint32 returns Uint32 from the global random source.
func Uint32() uint32 {
	return globalRand.Uint32()
}

// Uint64 returns two Uint32 calls arranged into a uint64.
func Uint64() uint64 {
	return uint64(Uint32())<<32 | uint64(Uint32())
}

// Perm returns Perm from the global random source.
func Perm(n int) []int {
	return globalRand.Perm(n)
}

// Read returns Read from the global random source.
func Read(p []byte) (n int, err error) {
	return globalRand.Read(p)
}

// Rand is a wrapper around the math/rand.Rand object.
type Rand struct {
	*rand.Rand
}

// New returns a new random source in the vein of math/rand.New.
func New(src Source) *Rand {
	return &Rand{
		Rand: rand.New(src),
	}
}

// SecureSeed seeds the random source with a cryptographically secure seed.
func (r *Rand) SecureSeed() {
	r.Seed(genSeed())
}

// Reseeder sets up a Reseeder on the random source to periodically reseed with a cryptographically secure seed.
func (r *Rand) Reseeder(d time.Duration) {
	r.SecureSeed()

	for range time.Tick(d) {
		r.SecureSeed()
	}
}

// Source is an alias for rand.Source.
type Source = rand.Source

// NewSource returns rand.NewSource(seed).
func NewSource(seed int64) Source {
	return rand.NewSource(seed)
}
