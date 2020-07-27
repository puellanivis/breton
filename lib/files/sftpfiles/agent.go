package sftpfiles

import (
	"net"
	"os"

	"golang.org/x/crypto/ssh/agent"
)

// Agent defines a connection to an ssh-agent through net.Conn.
type Agent struct {
	conn net.Conn
	agent.Agent
}

// Close implements io.Closer
func (a *Agent) Close() error {
	return a.conn.Close()
}

// GetAgent looks up the SSH_AUTH_SOCK environment variable, and creates a connection to it.
//
// If SSH_AUTH_SOCK is not set, then this returns both a nil Agent, and a nil error,
// as this is not an error condition.
//
// If SSH_AUTH_SOCK is set, any error attempting to connect to it will return an error, and a nil Agent.
func GetAgent() (*Agent, error) {
	sock := os.Getenv("SSH_AUTH_SOCK")
	if sock == "" {
		// No agent setup, so return no agent and no error.
		return nil, nil
	}

	raddr, err := net.ResolveUnixAddr("unix", sock)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUnix("unix", nil, raddr)
	if err != nil {
		return nil, err
	}

	return &Agent{
		conn:  conn,
		Agent: agent.NewClient(conn),
	}, nil
}
