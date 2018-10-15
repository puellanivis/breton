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
	hosts map[string]*Host
}

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
}

func init() {
	fs := &filesystem{
		hosts: make(map[string]*Host),
	}

	files.RegisterScheme(fs, "sftp", "scp")
}

func (fs *filesystem) getHost(uri *url.URL) *Host {
	fs.once.Do(fs.lazyInit)

	h := NewHost(uri)

	fs.mu.Lock()
	defer fs.mu.Unlock()

	key := h.Name()

	if h := fs.hosts[key]; h != nil {
		return h
	}

	_ = h.addAuths(fs.auths...)
	_, _ = h.SetHostKeyCallback(fs.knownhosts, nil)

	fs.hosts[key] = h

	return h
}

func (fs *filesystem) List(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	h := fs.getHost(uri)

	cl, err := h.Connect()
	if err != nil {
		return nil, &os.PathError{"connect", h.Name(), err}
	}

	fi, err := cl.ReadDir(uri.Path)
	if err != nil {
		return nil, &os.PathError{"readdir", uri.String(), err}
	}

	return fi, nil
}
