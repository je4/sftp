package ssh

import (
	"fmt"
	"github.com/goph/emperror"
	"github.com/op/go-logging"
	"golang.org/x/crypto/ssh"
	"net/url"
	"strings"
	"sync"
)

type ConnectionPool struct {
	// Protects access to fields below
	mu    sync.Mutex
	table map[string]*Connection
	log   *logging.Logger
}

func NewConnectionPool(log *logging.Logger) *ConnectionPool {
	return &ConnectionPool{
		mu:    sync.Mutex{},
		table: map[string]*Connection{},
		log:   log,
	}
}

func (cp *ConnectionPool) GetConnection(address *url.URL, config *ssh.ClientConfig) (*Connection, error) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	id := strings.ToLower(fmt.Sprintf("%s@%s", address.User.Username(), address.Host))

	conn, ok := cp.table[id]
	if ok {
		return conn, nil
	}
	var err error
	switch strings.ToLower(address.Scheme) {
	case "ssh":
	case "sftp":
		cp.log.Infof("new %s connection to %v", address.Scheme, id)
		conn, err = NewConnection(address, config, cp.log)
	default:
		return nil, emperror.Wrapf(err, "invalid scheme %s in %s", address.Scheme, address.String())
	}
	if err != nil {
		return nil, emperror.Wrapf(err, "cannot open ssh connection")
	}
	cp.table[id] = conn
	return conn, nil
}
