// Package httpfiles implements the "http:" and "https:" URL schemes.
package httpfiles

import (
	"errors"
	"net/http"
	"net/url"
	"os"

	"github.com/puellanivis/breton/lib/files"
)

type handler struct{}

var schemes = map[string]string{
	"http":  "80",
	"https": "443",
}

func init() {
	var schemeList []string

	for scheme := range schemes {
		schemeList = append(schemeList, scheme)
	}

	files.RegisterScheme(handler{}, schemeList...)
}

func elideDefaultPort(uri *url.URL) *url.URL {
	port := uri.Port()
	if port == "" {
		return uri
	}

	/* elide default ports  */
	if defport := schemes[uri.Scheme]; defport == port {
		u := *uri
		u.Host = uri.Hostname()
		return &u
	}

	return uri
}

func getErr(resp *http.Response) error {
	switch resp.StatusCode {
	case http.StatusOK, http.StatusNoContent:
		return nil
	case http.StatusUnauthorized, http.StatusForbidden:
		return os.ErrPermission
	case http.StatusNotFound:
		return os.ErrNotExist
	}

	return errors.New(resp.Status)
}
