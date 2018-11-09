package sftpfiles

import (
	"github.com/puellanivis/breton/lib/files"

	"golang.org/x/crypto/ssh"
)

func noopOption() files.Option {
	return func(_ files.File) (files.Option, error) {
		return noopOption(), nil
	}
}

func withAuths(auths []ssh.AuthMethod) files.Option {
	type authSetter interface {
		SetAuths([]ssh.AuthMethod) []ssh.AuthMethod
	}

	return func(f files.File) (files.Option, error) {
		h, ok := f.(authSetter)
		if !ok {
			return noopOption(), nil
		}

		save := h.SetAuths(auths)
		return withAuths(save), nil
	}
}

// WithAuth includes an arbitrary ssh.AuthMethod to be used for authentication during the ssh.Dial.
func WithAuth(auth ssh.AuthMethod) files.Option {
	type authAdder interface {
		AddAuth(ssh.AuthMethod) []ssh.AuthMethod
	}

	return func(f files.File) (files.Option, error) {
		h, ok := f.(authAdder)
		if !ok {
			return noopOption(), nil
		}

		save := h.AddAuth(auth)
		return withAuths(save), nil
	}
}

// IgnoreHostKeys specifies whether the ssh.Dial should ignore host keys during connection. Using this is insecure!
//
// Setting this to true will override any existing WithHostKey option, unless it is later turned off.
func IgnoreHostKeys(state bool) files.Option {
	type hostkeyIgnorer interface {
		IgnoreHostKeys(bool) bool
	}

	return func(f files.File) (files.Option, error) {
		h, ok := f.(hostkeyIgnorer)
		if !ok {
			return noopOption(), nil
		}

		save := h.IgnoreHostKeys(state)
		return IgnoreHostKeys(save), nil
	}
}

func withHostKeyCallback(cb ssh.HostKeyCallback, algos []string) files.Option {
	type hostkeySetter interface {
		SetHostKeyCallback(ssh.HostKeyCallback, []string) (ssh.HostKeyCallback, []string)
	}

	return func(f files.File) (files.Option, error) {
		h, ok := f.(hostkeySetter)
		if !ok {
			return noopOption(), nil
		}

		saveHK, saveAlgos := h.SetHostKeyCallback(cb, algos)
		return withHostKeyCallback(saveHK, saveAlgos), nil
	}
}

// WithHostKey defines an expected host key from the authorized key format specified in the sshd(8) man page.
//
// i.e. ssh-keytype BASE64BLOB string-comment
//
// If the IgnoreHostKeys option has been set, then this option will be ignored.
func WithHostKey(hostkey []byte) files.Option {
	type hostkeySetter interface {
		SetHostKeyCallback(ssh.HostKeyCallback) ssh.HostKeyCallback
	}

	key, _, _, _, err := ssh.ParseAuthorizedKey(hostkey)
	if err != nil {
		return func(_ files.File) (files.Option, error) {
			return nil, err
		}
	}

	return withHostKeyCallback(ssh.FixedHostKey(key), []string{key.Type()})
}
