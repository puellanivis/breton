package aboutfiles

import (
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/puellanivis/breton/lib/files"
	"github.com/puellanivis/breton/lib/files/wrapper"
)

type schemeList struct{}

func (schemeList) ReadAll() ([]byte, error) {
	schemes := files.RegisteredSchemes()

	var lines []string

	for _, scheme := range schemes {
		uri := &url.URL{
			Scheme: scheme,
		}

		lines = append(lines, uri.String())
	}

	return []byte(strings.Join(append(lines, ""), "\n")), nil
}

func (schemeList) ReadDir() ([]os.FileInfo, error) {
	schemes := files.RegisteredSchemes()

	var infos []os.FileInfo

	for _, scheme := range schemes {
		uri := &url.URL{
			Scheme: scheme,
		}

		infos = append(infos, wrapper.NewInfo(uri, 0, time.Now()))
	}

	return infos, nil
}
