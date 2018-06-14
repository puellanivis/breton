package rest

import (
	"net/http"
)

type Request struct {
	Path   []string
	Values map[string]interface{}

	*http.Request
}
