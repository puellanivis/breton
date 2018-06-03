package mapreduce

import (
	"context"
	"testing"
	"math/rand"
	"time"
)

type TestMR struct{
	ranges []Range
}

func (mr *TestMR) Map(ctx context.Context, in interface{}) (out interface{}, err error) {
	rng := in.(Range)

	<-time.After(time.Duration(rand.Intn(int(5 * time.Second))))

	return rng, nil
}

func (mr *TestMR) Reduce(ctx context.Context, in interface{}) error {
	rng := in.(Range)

	mr.ranges = append(mr.ranges, rng)

	return nil
}

func TestEngine(t *testing.T) {
	rng := &Range{
		Start: 42,
		End: 69,
	}

	DefaultThreadCount = 16

	mr := &TestMR{}

	errch := engine(context.Background(), mr, mr, *rng)
	for err := range errch {
		t.Error(err)
	}

	t.Log(mr.ranges)
}
