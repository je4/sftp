package ssh

import (
	"github.com/goph/emperror"
	"github.com/op/go-logging"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
)

type Connection struct {
	client               *ssh.Client
	config               *ssh.ClientConfig
	address              string
	log                  *logging.Logger
	concurrency          int
	maxClientConcurrency int
	maxPacketSize        int
}

func NewConnection(address, user string, config *ssh.ClientConfig, concurrency, maxClientConcurrency, maxPacketSize int, log *logging.Logger) (*Connection, error) {
	// create copy of config with user
	newConfig := &ssh.ClientConfig{
		Config:            config.Config,
		User:              user,
		Auth:              config.Auth,
		HostKeyCallback:   config.HostKeyCallback,
		BannerCallback:    config.BannerCallback,
		ClientVersion:     config.ClientVersion,
		HostKeyAlgorithms: config.HostKeyAlgorithms,
		Timeout:           config.Timeout,
	}

	sc := &Connection{
		client:               nil,
		log:                  log,
		config:               newConfig,
		address:              address,
		concurrency:          concurrency,
		maxClientConcurrency: maxClientConcurrency,
		maxPacketSize:        maxPacketSize,
	}
	// connect
	if err := sc.Connect(); err != nil {
		return nil, emperror.Wrapf(err, "cannot connect to %s@%s", user, address)
	}
	return sc, nil
}

func (sc *Connection) Connect() error {
	var err error
	sc.client, err = ssh.Dial("tcp", sc.address, sc.config)
	if err != nil {
		return emperror.Wrapf(err, "unable to connect to %v", sc.address)
	}

	return nil
}

func (sc *Connection) Close() {
	sc.client.Close()
}

func (sc *Connection) GetSFTPClient() (*sftp.Client, error) {
	sftpclient, err := sftp.NewClient(sc.client, sftp.MaxPacket(sc.maxPacketSize), sftp.MaxConcurrentRequestsPerFile(sc.maxClientConcurrency))
	if err != nil {
		sc.log.Infof("cannot get sftp subsystem - reconnecting to %s@%s", sc.client.User(), sc.address)
		if err := sc.Connect(); err != nil {
			return nil, emperror.Wrapf(err, "cannot connect with ssh to %s@%s", sc.client.User(), sc.address)
		}
		sftpclient, err = sftp.NewClient(sc.client)
		if err != nil {
			return nil, emperror.Wrapf(err, "cannot create sftp client on %s@%s", sc.client.User(), sc.address)
		}
	}
	return sftpclient, nil
}

func (sc *Connection) ReadFile(path string, w io.Writer) (int64, error) {
	sftpclient, err := sc.GetSFTPClient()
	if err != nil {
		return 0, emperror.Wrap(err, "unable to create SFTP session")
	}
	defer sftpclient.Close()

	r, err := sftpclient.Open(path)
	if err != nil {
		return 0, emperror.Wrapf(err, "cannot open remote file %s", path)
	}
	defer r.Close()

	written, err := r.WriteTo(w) // io.Copy(w, r)
	if err != nil {
		return 0, emperror.Wrap(err, "cannot copy data")
	}
	return written, nil
}

func (sc *Connection) WriteFile(path string, r io.Reader) (int64, error) {
	sftpclient, err := sc.GetSFTPClient()
	if err != nil {
		return 0, emperror.Wrap(err, "unable to create SFTP session")
	}
	defer sftpclient.Close()

	w, err := sftpclient.Create(path)
	if err != nil {
		return 0, emperror.Wrapf(err, "cannot create remote file %s", path)
	}

	written, err := w.ReadFromWithConcurrency(r, sc.concurrency) // io.Copy(w, r)
	if err != nil {
		return 0, emperror.Wrap(err, "cannot copy data")
	}
	return written, nil
}
