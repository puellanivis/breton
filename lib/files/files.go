// Package files implements an abstraction of accessing data via URL/URIs.
//
//
// The basic usage of the library is fairly straight-forward:
//	import (
//		"context"
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
//	defer r.Close()
//	io.Copy(os.Stdout, r)	// copy the response.Body from the HTTP request to stdout.
//
//	w, err := files.Create(ctx, "clipboard:")
//	if err != nil {
//		return err	// will return `os.IsNotExist(err) == true` if not available.
//	}
//	defer w.Close()
//	w.Write([]byte("write this string to the OS clipboard if available"))
//
//	fi, err := files.List(ctx, "home:")
//	if err != nil {
//		return err
//	}
//	for _, info := range fi {
//		// will print each filename and size listed in the user.Current().HomeDir
//		fmt.Println("%s %d", info.Name(), info.Size())
//	}
//
// Additionally, there are some helper functions which simply read/write a whole byte slice.
//
//	b, err := files.Read(ctx, "source")
//	if err != nil {
//		return err
//	}
//
//	err := files.Write(ctx, "destination", b)
//	if err != nil {
//		// will return io.ErrShortWrite if it does not write all of the buffer.
//		return err
//	}
//
// Or use io.ReaderCloser/io.WriteCloser as a source/destination
//
//	// discard and close all data on the io.ReadCloser
//	if err := files.Discard(io.ReadCloser); err != nil {
//		return err
//	}
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

// File defines an interface that abstracts away the concept of files to allow for multiple types of Scheme implementations, and not just local filesystem files.
type File interface {
	io.Closer
	Name() string
	Stat() (os.FileInfo, error)
}

// Reader defines a files.File that is available as an io.ReadSeeker
type Reader interface {
	File
	io.ReadSeeker
}

// Writer defines a files.File that is available as an io.Writer as well as supporting a Sync() function.
type Writer interface {
	File
	io.Writer
	Sync() error
}
