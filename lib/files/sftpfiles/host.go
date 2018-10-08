package sftpfiles

import (
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

func (h *host) GetClient() (*sftp.Client, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.cl != nil {
		return h.cl, nil
	}

	uri := h.uri

	if uri.Port() == "" {
		clone := *uri
		uri = &clone

		uri.Host = uri.Host + ":22"
	}

	hk := h.hostkey
	if h.ignoreHostkey {
		hk = ssh.InsecureIgnoreHostKey()
	}

	auths := h.cloneAuths()

	var user string

	switch {
	case uri.User != nil:
		user = uri.User.Username()

		if pw, ok := uri.User.Password(); ok {
			auths = append(auths, ssh.Password(pw))
		}

	default:
		user = username
	}

	conn, err := ssh.Dial("tcp", uri.Host, &ssh.ClientConfig{
		User:              user,
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

// AddAuths adds the given ssh.AuthMethod to the authorization methods for the host, and return the previous value.
func (h *host) AddAuths(auth ssh.AuthMethod) []ssh.AuthMethod {
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
