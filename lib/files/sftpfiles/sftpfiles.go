package sftpfiles

import (
	"context"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"sync"

	"github.com/puellanivis/breton/lib/files"

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

	var home string

	if u, err := user.Current(); err == nil {
		home = u.HomeDir
	}

	if home == "" {
		if h, err := os.UserHomeDir(); err == nil {
			home = h
		}
	}

	if home == "" {
		// couldn’t find a home directory, just give up trying to import known_hosts
		return
	}

	filename := filepath.Join(home, ".ssh", "known_hosts")

	if cb, err := knownhosts.New(filename); err == nil {
		fs.knownhosts = cb
	}
}

func init() {
	files.RegisterScheme(&filesystem{}, "sftp", "scp")
}

func (fs *filesystem) getHost(uri *url.URL) (*Host, *url.URL) {
	fs.once.Do(fs.lazyInit)

	h := NewHost(uri)
	key := h.Name()

	fs.mu.Lock()
	defer fs.mu.Unlock()

	if h := fs.hosts[key]; h != nil {
		return h, h.getPath(uri)
	}

	_ = h.addAuths(fs.auths...)
	_, _ = h.SetHostKeyCallback(fs.knownhosts, nil)

	if fs.hosts == nil {
		fs.hosts = make(map[string]*Host)
	}

	fs.hosts[key] = h

	return h, h.getPath(uri)
}

func (fs *filesystem) ReadDir(ctx context.Context, uri *url.URL) ([]os.FileInfo, error) {
	h, u := fs.getHost(uri)

	cl, err := h.Connect()
	if err != nil {
		return nil, &os.PathError{
			Op:   "connect",
			Path: h.Name(),
			Err:  err,
		}
	}

	fi, err := cl.ReadDir(uri.Path)
	if err != nil {
		return nil, &os.PathError{
			Op:   "readdir",
			Path: u.String(),
			Err:  err,
		}
	}

	return fi, nil
}
