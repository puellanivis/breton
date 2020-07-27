package sftpfiles

import (
	"errors"
	"net/url"
	"os/user"
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

	uri  *url.URL
	name string

	auths []ssh.AuthMethod

	ignoreHostkey bool
	hostkey       ssh.HostKeyCallback
	hostkeyAlgos  []string
}

var (
	userInit    sync.Once
	defaultUser *url.Userinfo
)

func getUser() *url.Userinfo {
	userInit.Do(func() {
		u, err := user.Current()
		if err != nil {
			return
		}

		defaultUser = url.User(u.Username)
	})

	return defaultUser
}

// NewHost returns a Host defined for a specific host/user based on a given URL.
// No connection is made, and no authentication or hostkey validation is defined.
func NewHost(uri *url.URL) *Host {
	var auths []ssh.AuthMethod

	user := getUser()
	if uri.User != nil {
		user = url.User(uri.User.Username())

		if pw, ok := uri.User.Password(); ok {
			auths = append(auths, ssh.Password(pw))
		}
	}

	uri = &url.URL{
		Scheme: "ssh",
		Host:   uri.Host,
		User:   user,
	}

	if uri.Port() == "" {
		uri.Host += ":22"
	}

	return &Host{
		uri:   uri,
		name:  uri.String(),
		auths: auths,
	}
}

// Name returns an identifying name of the Host composed of the authority section of the URL: //user[:pass]@hostname:port
func (h *Host) Name() string {
	return h.name
}

func (h *Host) getPath(uri *url.URL) *url.URL {
	u := *uri
	u.Host = h.uri.Host
	u.User = h.uri.User

	return &u
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

func (h *Host) client() *sftp.Client {
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

	return h.client()
}

// Connect either returns the currently connected Client, or makes a new connection based on Host.
func (h *Host) Connect() (*sftp.Client, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if cl := h.client(); cl != nil {
		return cl, nil
	}

	hk := h.hostkey
	if h.ignoreHostkey {
		hk = ssh.InsecureIgnoreHostKey()
	}

	if hk == nil {
		return nil, errors.New("no hostkey validation defined")
	}

	conn, err := ssh.Dial("tcp", h.uri.Host, &ssh.ClientConfig{
		User:              h.uri.User.Username(),
		Auth:              h.cloneAuths(),
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
	return append([]ssh.AuthMethod(nil), h.auths...)
}

// addAuths is an internal convenience func to add any number of auths.
func (h *Host) addAuths(auths ...ssh.AuthMethod) []ssh.AuthMethod {
	h.mu.Lock()
	defer h.mu.Unlock()

	clone := h.cloneAuths()

	if len(auths) < 1 {
		return clone
	}

	return h.setAuths(append(clone, auths...))
}

// AddAuth adds the given ssh.AuthMethod to the authorization methods for the Host, and return the previous value.
func (h *Host) AddAuth(auth ssh.AuthMethod) []ssh.AuthMethod {
	h.mu.Lock()
	defer h.mu.Unlock()

	return h.setAuths(append(h.cloneAuths(), auth))
}

// setAuths sets the slice of ssh.AuthMethod on the Host, and returns the previous value.
// It must be called under lock.
func (h *Host) setAuths(auths []ssh.AuthMethod) []ssh.AuthMethod {
	save := h.auths

	h.auths = auths

	return save
}

// SetAuths sets the slice of ssh.AuthMethod on the Host, and returns the previous value.
func (h *Host) SetAuths(auths []ssh.AuthMethod) []ssh.AuthMethod {
	h.mu.Lock()
	defer h.mu.Unlock()

	return h.setAuths(append([]ssh.AuthMethod(nil), auths...))
}

// IgnoreHostKeys sets a flag that Host should ignore Host keys when connecting.
// THIS IS INSECURE.
func (h *Host) IgnoreHostKeys(state bool) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	save := h.ignoreHostkey

	h.ignoreHostkey = state

	return save
}

// SetHostKeyCallback sets the current hostkey callback for the Host, and returns the previous value.
func (h *Host) SetHostKeyCallback(cb ssh.HostKeyCallback, algos []string) (ssh.HostKeyCallback, []string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	saveHK, saveAlgos := h.hostkey, h.hostkeyAlgos

	h.hostkey = cb
	h.hostkeyAlgos = algos

	return saveHK, saveAlgos
}
