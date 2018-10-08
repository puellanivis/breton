package sftpfiles

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sync"

	"github.com/puellanivis/breton/lib/files"
	"github.com/puellanivis/breton/lib/os/user"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type filesystem struct {
	agent      *Agent
	auths      []ssh.AuthMethod
	knownhosts ssh.HostKeyCallback

	mu    sync.Mutex
	hosts map[string]*host
}

var username string

func init() {
	fs := &filesystem{}

	if agent, err := GetAgent(); err == nil && agent != nil {
		fs.agent = agent
		fs.auths = append(fs.auths, ssh.PublicKeysCallback(agent.Signers))
	}

	if home, err := user.CurrentHomeDir(); err == nil {
		filename := filepath.Join(home, ".ssh", "known_hosts")

		if cb, err := knownhosts.New(filename); err == nil {
			fs.knownhosts = cb
		}
	}

	if name, err := user.CurrentUsername(); err == nil {
		username = name
	}

	files.RegisterScheme(fs, "sftp", "scp")
}

func hashURL(uri *url.URL) string {
	if uri.User == nil {
		return uri.Host
	}

	return fmt.Sprintf("%s@%s", uri.User, uri.Host)
}

func (fs *filesystem) getHost(uri *url.URL) *host {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	hash := hashURL(uri)

	if h, ok := fs.hosts[hash]; ok {
		return h
	}

	h := &host{
		uri:     uri,
		auths:   fs.auths,
		hostkey: fs.knownhosts,
	}

	fs.hosts[hash] = h

	return h
}

func (fs *filesystem) Create(ctx context.Context, uri *url.URL) (files.Writer, error) {
	return nil, &os.PathError{"create", uri.String(), errors.New("not implemented")}
}

func (fs *filesystem) List(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	return nil, &os.PathError{"readdir", uri.String(), errors.New("not implemented")}
}
