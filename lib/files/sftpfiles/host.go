package sftpfiles

import (
	"errors"
	"net/url"
	"sync"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// Host defines a set of connection settings to a specific host/user combination,
// and manages a common SFTP connection to that host with those credentials.
type Host struct {
	mu   sync.Mutex
	conn *ssh.Client
	cl   *sftp.Client

	uri *url.URL

	auths []ssh.AuthMethod

	ignoreHostkey bool
	hostkey       ssh.HostKeyCallback
	hostkeyAlgos  []string
}

// NewHost returns a Host defined for a specific host/user based on a given URL.
// No connection is made, and no authentication or hostkey validation is defined.
func NewHost(uri *url.URL) *Host {
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

	return &Host{
		uri: uri,
	}
}

// Name returns an identifying name of the Host composed of the authority section of the URL: //user[:pass]@hostname:port
func (h *Host) Name() string {
	return h.uri.String()
}

func (h *Host) close() error {
	if h.cl == nil {
		return nil
	}

	err := h.cl.Close()
	if err2 := h.conn.Close(); err == nil {
		err = err2
	}

	h.cl, h.conn = nil, nil

	return err
}

// Close closes and invalidates the Host's current connection.
func (h *Host) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	return h.close()
}

func (h *Host) getClient() *sftp.Client {
	if h.cl == nil {
		return nil
	}

	if _, err := h.cl.Getwd(); err != nil {
		// We cannot get the current working directory,
		// So, invalidate our connections, and return nil.
		_ = h.close()

		return nil
	}

	return h.cl
}

// GetClient returns the currently connected Client connected to by the Host.
// It returns nil if the Host is not currently connected.
func (h *Host) GetClient() *sftp.Client {
	h.mu.Lock()
	defer h.mu.Unlock()

	return h.getClient()
}

// ConnectClient either returns the currently connected Client, or makes a new connection based on Host.
func (h *Host) ConnectClient() (*sftp.Client, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if cl := h.getClient(); cl != nil {
		return cl, nil
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

func (h *Host) cloneAuths() []ssh.AuthMethod {
	return append([]ssh.AuthMethod{}, h.auths...)
}

// AddAuth adds the given ssh.AuthMethod to the authorization methods for the Host, and return the previous value.
func (h *Host) AddAuth(auth ssh.AuthMethod) []ssh.AuthMethod {
	return h.SetAuths(append(h.cloneAuths(), auth))
}

// SetAuths sets the slice of ssh.AuthMethod on the Host, and returns the previous value.
func (h *Host) SetAuths(auths []ssh.AuthMethod) []ssh.AuthMethod {
	save := h.auths

	h.auths = auths

	return save
}

// IgnoreHostKeys sets a flag that Host should ignore Host keys when connecting.
// THIS IS INSECURE.
func (h *Host) IgnoreHostKeys(state bool) bool {
	save := h.ignoreHostkey

	h.ignoreHostkey = state

	return save
}

// SetHostKeyCallback sets the current hostkey callback for the Host, and returns the previous value.
func (h *Host) SetHostKeyCallback(cb ssh.HostKeyCallback, algos []string) (ssh.HostKeyCallback, []string) {
	saveHK, saveAlgos := h.hostkey, h.hostkeyAlgos

	h.hostkey = cb
	h.hostkeyAlgos = algos

	return saveHK, saveAlgos
}
