package ssh

import (
	"github.com/goph/emperror"
	"github.com/op/go-logging"
	"golang.org/x/crypto/ssh"
	"net/url"
)

type Connection struct {
	Client  *ssh.Client
	config  *ssh.ClientConfig
	Address *url.URL
	Log     *logging.Logger
}

func NewConnection(address *url.URL, config *ssh.ClientConfig, log *logging.Logger) (*Connection, error) {
	// create copy of config with user
	newConfig := &ssh.ClientConfig{
		Config:            config.Config,
		User:              address.User.Username(),
		Auth:              config.Auth,
		HostKeyCallback:   config.HostKeyCallback,
		BannerCallback:    config.BannerCallback,
		ClientVersion:     config.ClientVersion,
		HostKeyAlgorithms: config.HostKeyAlgorithms,
		Timeout:           config.Timeout,
	}

	sc := &Connection{
		Client:  nil,
		Log:     log,
		config:  newConfig,
		Address: address,
	}
	// connect
	if err := sc.Connect(); err != nil {
		return nil, emperror.Wrapf(err, "cannot connect to %s", address.String())
	}
	return sc, nil
}

func (sc *Connection) Connect() error {
	var err error
	sc.Client, err = ssh.Dial("tcp", sc.Address.Host, sc.config)
	if err != nil {
		return emperror.Wrapf(err, "unable to connect to %v", sc.Address)
	}

	return nil
}

func (sc *Connection) Close() {
	sc.Client.Close()
}

/*
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
*/
