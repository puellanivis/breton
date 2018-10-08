package sftpfiles

import (
	"github.com/puellanivis/breton/lib/files"

	"golang.org/x/crypto/ssh"
)

func noopOption() (files.Option, error) {
	return func(_ files.File) (files.Option, error) {
		return noopOption()
	}, nil
}

func withAuths(auths []ssh.AuthMethod) files.Option {
	type authSetter interface {
		SetAuths([]ssh.AuthMethod) []ssh.AuthMethod
	}

	return func(f files.File) (files.Option, error) {
		h, ok := f.(authSetter)
		if !ok {
			return noopOption()
		}

		save := h.SetAuths(auths)
		return withAuths(save), nil
	}
}

func WithAuth(auth ssh.AuthMethod) files.Option {
	type authAdder interface {
		AddAuth(ssh.AuthMethod) []ssh.AuthMethod
	}

	return func(f files.File) (files.Option, error) {
		h, ok := f.(authAdder)
		if !ok {
			return noopOption()
		}

		save := h.AddAuth(auth)
		return withAuths(save), nil
	}
}

func IgnoreHostKeys(state bool) files.Option {
	type ignorer interface {
		IgnoreHostKeys(bool) bool
	}

	return func(f files.File) (files.Option, error) {
		h, ok := f.(ignorer)

		if !ok {
			return noopOption()
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
			return noopOption()
		}

		saveHK, saveAlgos := h.SetHostKeyCallback(cb, algos)
		return withHostKeyCallback(saveHK, saveAlgos), nil
	}
}

func WithHostKey(hostkey []byte) files.Option {
	type hostkeySetter interface {
		SetHostKeyCallback(ssh.HostKeyCallback) ssh.HostKeyCallback
	}

	key, _, _, _, err := ssh.ParseAuthorizedKey(hostkey)
	if err != nil {
		return func(f files.File) (files.Option, error) {
			return nil, err
		}
	}

	return withHostKeyCallback(ssh.FixedHostKey(key), []string{key.Type()})
}
