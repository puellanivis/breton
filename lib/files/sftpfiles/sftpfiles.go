package sftpfiles

import (
	"context"
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
	once sync.Once

	agent      *Agent
	auths      []ssh.AuthMethod
	knownhosts ssh.HostKeyCallback

	mu    sync.Mutex
	hosts map[string]*host
}

var username string

func (fs *filesystem) lazyInit() {
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
}

func init() {
	fs := &filesystem{
		hosts: make(map[string]*host),
	}

	files.RegisterScheme(fs, "sftp", "scp")
}

func (fs *filesystem) getHost(uri *url.URL) *host {
	fs.once.Do(fs.lazyInit)

	uri = &url.URL{
		Host: uri.Host,
		User: uri.User,
	}

	if uri.Port() == "" {
		uri.Host += ":22"
	}

	if uri.User == nil {
		uri.User = url.User(username)
	}

	fs.mu.Lock()
	defer fs.mu.Unlock()

	key := uri.String()

	if h := fs.hosts[key]; h != nil {
		return h
	}

	h := &host{
		uri:     uri,
		auths:   append([]ssh.AuthMethod{}, fs.auths...),
		hostkey: fs.knownhosts,
	}

	fs.hosts[key] = h

	return h
}

func (fs *filesystem) List(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	h := fs.getHost(uri)

	cl, err := h.ConnectClient()
	if err != nil {
		return nil, &os.PathError{"connect", uri.String(), err}
	}

	fi, err := cl.ReadDir(uri.Path)
	if err != nil {
		return nil, &os.PathError{"readdir", uri.String(), err}
	}

	return fi, nil
}
