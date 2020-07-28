// Package files implements an abstraction of accessing data via URL/URIs.
//
//
// The basic usage of the library is fairly straight-forward:
//	import (
//		"context"
//		"fmt"
//		"io"
//		"os"
//
//		"github.com/puellanivis/breton/lib/files"
//		_ "github.com/puellanivis/breton/lib/files/plugins"
//	)
//
//	ctx = context.Background()
//
//	r, err := files.Open(ctx, "http://www.example.com")
//	if err != nil {
//		return err
//	}
//	// copy the response.Body from the HTTP request to stdout.
//	if _, err := files.Copy(ctx, os.Stdout, r); err != nil {
//		return err
//	}
//	r.Close()
//
//	w, err := files.Create(ctx, "clipboard:")
//	if err != nil {
//		// will return `os.IsNotExist(err) == true` if not available.
//		return err
//	}
//	fmt.Fprint(w, "write this string to the OS clipboard if available")
//	w.Close()
//
//	fi, err := files.List(ctx, "home:")
//	if err != nil {
//		return err
//	}
//	for _, info := range fi {
//		// will print each filename and size listed in the user.Current().HomeDir
//		fmt.Printf("%s %d\n", info.Name(), info.Size())
//	}
//
// Additionally, there are some helper functions which simply read/write a whole byte slice.
//
//	b, err := files.Read(ctx, "source")
//	if err != nil {
//		return err
//	}
//	// no need to close here.
//
//	err := files.Write(ctx, "destination", b)
//	if err != nil {
//		// will return io.ErrShortWrite if it does not write all of the buffer.
//		return err
//	}
//	// no need to close here.
//
// Or use io.ReaderCloser/io.WriteCloser as a source/destination
//
//	// discard and close all data on the io.ReadCloser
//	if err := files.Discard(io.ReadCloser); err != nil {
//		return err
//	}
//	// no need to close here.
//
//	// read all the data on the io.ReadCloser into a byte-slice, then close the source.
//	b, err := files.ReadFrom(io.ReadCloser)
//	if err != nil {
//		return err
//	}
//	// no need to close here.
//
//	// write all the data from the byte-slice into the io.WriteCloser, then close the source.
//	if err := files.WriteTo(io.WriteCloser, b); err != nil {
//		return err
//	}
//	// no need to close here.
//
package files

import (
	"io"
	"os"
)

// File defines an interface that abstracts the central core concepts of files, for broad implementation.
type File interface {
	io.Closer
	Name() string
	Stat() (os.FileInfo, error)
}

// Reader defines an extension interface on files.File that is also an io.Reader.
type Reader interface {
	File
	io.Reader
}

// SeekReader defines an extension interface on files.Reader that is also an io.Seeker
type SeekReader interface {
	Reader
	io.Seeker
}

// Writer defines an extention interface on files.File that is also an io.Writer.
type Writer interface {
	File
	io.Writer
}

// SyncWriter defines an extension interface on files.Writer that also supports Sync().
type SyncWriter interface {
	Writer
	Sync() error
}
