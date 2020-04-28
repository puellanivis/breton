// Package plugins is imported for side-effects and includes default files scheme plugins.
package plugins

import (
	// this includes all default plugins for lib/files
	_ "github.com/puellanivis/breton/lib/files/about"
	_ "github.com/puellanivis/breton/lib/files/cachefiles"
	_ "github.com/puellanivis/breton/lib/files/clipboard"
	_ "github.com/puellanivis/breton/lib/files/datafiles"
	_ "github.com/puellanivis/breton/lib/files/home"
	_ "github.com/puellanivis/breton/lib/files/httpfiles"
	_ "github.com/puellanivis/breton/lib/files/socketfiles"
)
