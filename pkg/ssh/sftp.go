package ssh

import (
	"github.com/goph/emperror"
	"github.com/je4/sftp/v2/pkg/stream"
	"github.com/op/go-logging"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"time"
)

type SFTP struct {
	config               *ssh.ClientConfig
	log                  *logging.Logger
	pool                 *ConnectionPool
	concurrency          int
	maxClientConcurrency int
	maxPacketSize        int
	rsc                  *stream.ReadStreamQueue
}

func NewSFTP(PrivateKey []string, Password, KnownHosts string, concurrency, maxClientConcurrency, maxPacketSize int, rsc *stream.ReadStreamQueue, log *logging.Logger) (*SFTP, error) {
	var signer []ssh.Signer

	readStreamQueue, err := stream.NewReadStreamQueue(rsc)
	if err != nil {
		return nil, emperror.Wrap(err, "cannot create ReadStreamQueue")
	}

	sftp := &SFTP{
		log: log,
		config: &ssh.ClientConfig{
			Auth:            []ssh.AuthMethod{},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		},
		pool:                 NewConnectionPool(log),
		concurrency:          concurrency,
		maxClientConcurrency: maxClientConcurrency,
		maxPacketSize:        maxPacketSize,
		rsc:                  readStreamQueue,
	}

	for _, pk := range PrivateKey {
		key, err := ioutil.ReadFile(pk)
		if err != nil {
			return nil, emperror.Wrapf(err, "cannot read private key file %s")
		}
		// Create the Signer for this private key.
		s, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, emperror.Wrapf(err, "unable to parse private key %v", string(key))
		}
		signer = append(signer, s)
	}
	if len(signer) > 0 {
		sftp.config.Auth = append(sftp.config.Auth, ssh.PublicKeys(signer...))
	}
	if KnownHosts != "" {
		hostKeyCallback, err := knownhosts.New(KnownHosts)
		if err != nil {
			return nil, emperror.Wrapf(err, "could not create hostkeycallback function for %s", KnownHosts)
		}
		sftp.config.HostKeyCallback = hostKeyCallback
	}
	if Password != "" {
		sftp.config.Auth = append(sftp.config.Auth, ssh.Password(Password))
	}
	return sftp, nil
}

func (s *SFTP) GetConnection(address *url.URL) (*Connection, error) {
	return s.pool.GetConnection(address, s.config)
}

func (s *SFTP) Get(uri *url.URL, w io.Writer) (int64, error) {
	conn, err := s.GetConnection(uri)
	if err != nil {
		return 0, emperror.Wrapf(err, "unable to connect to %v with user %v", uri.String(), uri.User.Username())
	}
	sConn, err := NewSFTPConnection(conn, s.concurrency, s.maxClientConcurrency, s.maxPacketSize)
	if err != nil {
		return 0, emperror.Wrapf(err, "unable to create sftp connection for %s", uri.String())
	}
	written, err := sConn.ReadFile(uri.Path, w)
	if err != nil {
		return 0, emperror.Wrapf(err, "cannot read data from %v", uri.Path)
	}
	return written, nil
}

func (s *SFTP) GetFile(uri *url.URL, user string, target string) (int64, error) {
	f, err := os.Create(target)
	if err != nil {
		return 0, emperror.Wrapf(err, "cannot create file %s", target)
	}
	defer f.Close()
	return s.Get(uri, f)
}

func (s *SFTP) PutFile(uri *url.URL, source string) (int64, error) {
	f, err := os.Open(source)
	if err != nil {
		return 0, emperror.Wrapf(err, "cannot open file %s", source)
	}
	defer f.Close()
	return s.Put(uri, f)
}

func (s *SFTP) Put(uri *url.URL, r io.Reader) (int64, error) {
	conn, err := s.GetConnection(uri)
	if err != nil {
		return 0, emperror.Wrapf(err, "unable to connect to %v with user %v", uri.String(), uri.User.Username())
	}
	sConn, err := NewSFTPConnection(conn, s.concurrency, s.maxClientConcurrency, s.maxPacketSize)
	if err != nil {
		return 0, emperror.Wrapf(err, "unable to create sftp connection for %s", uri.String())
	}
	daReader := s.rsc.StartReader(r)
	start := time.Now()
	written, err := sConn.WriteFile(uri.Path, daReader)
	if err != nil {
		return written, err
	}
	since := time.Since(start)
	s.log.Infof("written: %dB %dns %.2fGB %.2fs %.2fMB/s\n",
		written, since,
		float64(written)/1000000000, float64(since)/float64(time.Second),
		(float64(written)/1000000)/(float64(since)/float64(time.Second)),
	)
	return written, err
}
