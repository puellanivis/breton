// Wraps the golang standard library, unless it's Windows, then use NetLookup instead of DomainLookup first.

// +build !windows

package user

import (
	"os/user"
)

type Group = user.Group
type User = user.User

func LookupGroup(name string) (*Group, error) {
	return user.LookupGroup(name)
}

func LookupGroupId(gid string) (*Group, error) {
	return user.LookupGroupId(gid)
}

func Current() (*User, error) {
	return user.Current()
}

func Lookup(username string) (*User, error) {
	return user.Lookup(username)
}

func LookupId(uid string) (*User, error) {
	return user.LookupId(uid)
}

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
