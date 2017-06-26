package plugins

import (
	// this includes all default plugins for lib/files
	_ "lib/files/about"
	_ "lib/files/cache"
	_ "lib/files/clipboard"
	_ "lib/files/data"
	_ "lib/files/home"
	_ "lib/files/http"
)
