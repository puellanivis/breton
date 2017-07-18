package rand

import (
	"math/rand"
	"time"
)

func ExpFloat64() float64 {
	return globalRand.ExpFloat64()
}

func Float32() float32 {
	return globalRand.Float32()
}

func Float64() float64 {
	return globalRand.Float64()
}

func Int() int {
	return globalRand.Int()
}

func Int31() int32 {
	return globalRand.Int31()
}

func Int31n(n int32) int32 {
	return globalRand.Int31n(n)
}

func Int63() int64 {
	return globalRand.Int63()
}

func Int63n(n int64) int64 {
	return globalRand.Int63n(n)
}

func Intn(n int) int {
	return globalRand.Intn(n)
}

func NormFloat64() float64 {
	return globalRand.NormFloat64()
}

func Perm(n int) []int {
	return globalRand.Perm(n)
}

func Read(p []byte) (n int, err error) {
	return globalRand.Read(p)
}

func Uint32() uint32 {
	return globalRand.Uint32()
}

func Uint64() uint64 {
	return uint64(Uint32())<<32 | uint64(Uint32())
}

type Rand struct {
	*rand.Rand
}

func New(src Source) *Rand {
	return &Rand{
		Rand: rand.New(src),
	}
}

func (r *Rand) SecureSeed() {
	r.Seed(genSeed())
}

func (r *Rand) Reseeder(d time.Duration) {
	r.SecureSeed()

	for range time.Tick(d) {
		r.SecureSeed()
	}
}

type Source interface {
	Int63() int64
	Seed(seed int64)
}

func NewSource(seed int64) Source {
	return rand.NewSource(seed)
}
