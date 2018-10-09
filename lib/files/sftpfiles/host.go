package sftpfiles

import (
	"errors"
	"net/url"
	"sync"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type host struct {
	mu   sync.Mutex
	conn *ssh.Client
	cl   *sftp.Client

	uri *url.URL

	auths []ssh.AuthMethod

	ignoreHostkey bool
	hostkey       ssh.HostKeyCallback
	hostkeyAlgos  []string
}

func (h *host) GetClient() *sftp.Client {
	h.mu.Lock()
	defer h.mu.Unlock()

	return h.cl
}

func (h *host) ConnectClient() (*sftp.Client, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.cl != nil {
		return h.cl, nil
	}

	hk := h.hostkey
	if h.ignoreHostkey {
		hk = ssh.InsecureIgnoreHostKey()
	}

	if hk == nil {
		return nil, errors.New("no hostkey validation defined")
	}

	auths := h.cloneAuths()

	if pw, ok := h.uri.User.Password(); ok {
		auths = append(auths, ssh.Password(pw))
	}

	conn, err := ssh.Dial("tcp", h.uri.Host, &ssh.ClientConfig{
		User:              h.uri.User.Username(),
		Auth:              auths,
		HostKeyCallback:   hk,
		HostKeyAlgorithms: h.hostkeyAlgos,
	})
	if err != nil {
		return nil, err
	}

	cl, err := sftp.NewClient(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	h.conn, h.cl = conn, cl

	return cl, nil
}

func (h *host) cloneAuths() []ssh.AuthMethod {
	return append([]ssh.AuthMethod{}, h.auths...)
}

// AddAuth adds the given ssh.AuthMethod to the authorization methods for the host, and return the previous value.
func (h *host) AddAuth(auth ssh.AuthMethod) []ssh.AuthMethod {
	return h.SetAuths(append(h.cloneAuths(), auth))
}

// SetAuths sets the slice of ssh.AuthMethod on the host, and returns the previous value.
func (h *host) SetAuths(auths []ssh.AuthMethod) []ssh.AuthMethod {
	save := h.auths

	h.auths = auths

	return save
}

func (h *host) IgnoreHostKeys(state bool) bool {
	save := h.ignoreHostkey

	h.ignoreHostkey = state

	return save
}

// SetHostKeyCallback sets the current hostkey callback for the host, and returns the previous value.
func (h *host) SetHostKeyCallback(cb ssh.HostKeyCallback, algos []string) (ssh.HostKeyCallback, []string) {
	saveHK, saveAlgos := h.hostkey, h.hostkeyAlgos

	h.hostkey = cb
	h.hostkeyAlgos = algos

	return saveHK, saveAlgos
}
