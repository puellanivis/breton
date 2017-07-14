// Wraps the golang standard library, unless it's Windows, then use NetLookup instead of DomainLookup first.

// +build !windows

package user

import (
	"os/user"
)

type Group struct{ user.Group }
type User struct{ user.User }

var (
	LookupGroup   = user.LookupGroup
	LookupGroupId = user.LookupGroupId
	Current       = user.Current
	Lookup        = user.Lookup
	LookupId      = user.LookupId
)

func CurrentHomeDir() (string, error) {
	me, err := user.Current()
	if err != nil {
		return "", err
	}

	return me.HomeDir, nil
}

func CurrentUsername() (string, error) {
	me, err := user.Current()
	if err != nil {
		return "", err
	}

	return me.Username, nil
}
