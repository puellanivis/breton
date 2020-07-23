// Package home implements a URL scheme "home:" which references files according to user home directories.
package home

import (
	"errors"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"sync"
)

type cache struct {
	mu sync.RWMutex

	cur   *user.User
	users map[string]*user.User
}

func (c *cache) current() (*user.User, error) {
	c.mu.RLock()
	u := c.cur
	c.mu.RUnlock()

	if u != nil {
		return u, nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if u := c.cur; u != nil {
		// Another thread already did the work.
		return u, nil
	}

	u, err := user.Current()
	if err != nil {
		return nil, err
	}

	c.cur = u

	if c.users == nil {
		c.users = make(map[string]*user.User)
	}

	c.users[u.Username] = u

	return u, nil
}

func (c *cache) lookup(username string) (*user.User, error) {
	c.mu.RLock()
	u := c.users[username]
	c.mu.RUnlock()

	if u != nil {
		return u, nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if u = c.users[username]; u != nil {
		// Another thread already did the work.
		return u, nil
	}

	u, err := user.Lookup(username)
	if err != nil {
		return nil, err
	}

	if c.users == nil {
		c.users = make(map[string]*user.User)
	}

	c.users[username] = u

	return u, nil
}

var users cache

// Filename takes a given url, and returns a filename that is an absolute path
// for the specific default user if home:filename, or a specific user if home://user@/filename.
func Filename(uri *url.URL) (string, error) {
	if uri.Host != "" {
		return "", os.ErrInvalid
	}

	path := uri.Path
	if path == "" {
		path = uri.Opaque
	}

	var base string

	switch uri.User {
	case nil:
		u, err := users.current()
		if err != nil {
			return "", err
		}

		base = u.HomeDir

	default:
		u, err := users.lookup(uri.User.Username())
		if err != nil {
			return "", err
		}

		base = u.HomeDir
	}

	if base == "" {
		return "", errors.New("could not find home directory")
	}

	return filepath.Join(base, path), nil
}
