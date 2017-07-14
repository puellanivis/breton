package files

import (
	"errors"
	"os"
)

type Option func(File) (Option, error)

// applyOptions is a helper function to apply a range of options on an os.File
func applyOptions(f File, opts []Option) error {
	for _, opt := range opts {
		_, err := opt(f)
		if err != nil {
			return err
		}
	}

	return nil
}

var ErrNotSupported = errors.New("option not supported")

func WithFileMode(mode os.FileMode) Option {
	type chmoder interface {
		Chmod(os.FileMode) error
	}

	return func(f File) (Option, error) {
		fi, err := f.Stat()
		if err != nil {
			return nil, err
		}

		save := fi.Mode()

		switch f := f.(type) {
		case chmoder:
			if err := f.Chmod(mode); err != nil {
				return nil, err
			}

		default:
			return nil, ErrNotSupported
		}

		return WithFileMode(save), nil
	}
}
