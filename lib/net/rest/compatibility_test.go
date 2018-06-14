package rest

import (
	"net/http"
	"testing"
)

func TestTypes(t *testing.T) {
	var f HandlerFunc
	var i Handler

	// a HandlerFunc must be a Handler
	i = f
	_ = i

	var h Handler

	// a ServeMux must be a Handler
	h = new(ServeMux)
	_ = h

	var hh http.Handler

	// a ServeMux must be an http.Handler
	hh = new(ServeMux)
	_ = hh
}
