package ssh

import (
	"fmt"
	"github.com/goph/emperror"
	"github.com/op/go-logging"
	"golang.org/x/crypto/ssh"
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

func (cp *ConnectionPool) GetConnection(address, user string, config *ssh.ClientConfig, concurrency, maxClientConcurrency, maxPackedSize int) (*Connection, error) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	id := strings.ToLower(fmt.Sprintf("%s@%s", user, address))

	conn, ok := cp.table[id]
	if ok {
		return conn, nil
	}
	var err error
	cp.log.Infof("new ssh connection to %v", id)
	conn, err = NewConnection(address, user, config, concurrency, maxClientConcurrency, maxPackedSize, cp.log)
	if err != nil {
		return nil, emperror.Wrapf(err, "cannot open ssh connection")
	}
	cp.table[id] = conn
	return conn, nil
}
