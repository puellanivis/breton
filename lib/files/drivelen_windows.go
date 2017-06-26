package files

import (
	"net/url"
)

// since we're on Windows, if the scheme is one or less chars
// then it either isn't set at all, or it's a drive letter.
const localDriveLength = 1
