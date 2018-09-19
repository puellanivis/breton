package mapreduce

import (
	"context"
	"strings"
	"testing"
)

func TestExample(t *testing.T) {
	mapper := MapFunc(func(ctx context.Context, in interface{}) (interface{}, error) {
		letters := in.([]string)

		var ucLetters []string

		for _, letter := range letters {
			ucLetters = append(ucLetters, strings.ToUpper(letter))
		}

		return ucLetters, nil
	})

	var collection [][]string

	reducer := ReduceFunc(func(ctx context.Context, in interface{}) error {
		ucLetters := in.([]string)

		collection = append(collection, ucLetters)

		return nil
	})

	mr := New(mapper, reducer)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	exampleInput := strings.Split("abcdefghijklmnopqrstuvwxyz", "")

	for err := range mr.Run(ctx, exampleInput) {
		t.Error(err)
	}

	t.Log(collection)
}
