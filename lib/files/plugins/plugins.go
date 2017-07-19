package plugins

import (
	// this includes all default plugins for lib/files
	_ "github.com/puellanivis/breton/lib/files/about"
	_ "github.com/puellanivis/breton/lib/files/cache"
	_ "github.com/puellanivis/breton/lib/files/clipboard"
	_ "github.com/puellanivis/breton/lib/files/data"
	_ "github.com/puellanivis/breton/lib/files/home"
	_ "github.com/puellanivis/breton/lib/files/http"
)
