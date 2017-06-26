package files

import (
	"context"
	"io"
	"io/ioutil"
)

func ReadAndClose(r io.ReadCloser) ([]byte, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return b, r.Close()
}

func Discard(r io.ReadCloser) error {
	if _, err := io.Copy(ioutil.Discard, r); err != nil {
		return err
	}

	return r.Close()
}

func ReadFile(ctx context.Context, filename string) ([]byte, error) {
	f, err := Open(ctx, filename)
	if err != nil {
		return nil, err
	}

	return ReadAndClose(f)
}
