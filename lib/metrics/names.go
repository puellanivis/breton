package metrics

import (
	"regexp"
)

var validName = regexp.MustCompile(`^[a-zA-Z_:][a-zA-Z0-9_:]*$`)
