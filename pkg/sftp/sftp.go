package sftp

import (
	"fmt"
	"github.com/goph/emperror"
	xssh "github.com/je4/sftp/v2/pkg/ssh"
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
	pool                 *xssh.ConnectionPool
	concurrency          int
	maxClientConcurrency int
	maxPacketSize        int
	queue                []TransferQueueEntry
}

func NewSFTP(PrivateKey []string, Password, KnownHosts string, concurrency, maxClientConcurrency, maxPacketSize int, log *logging.Logger, queue ...TransferQueueEntry) (*SFTP, error) {
	var signer []ssh.Signer

	sftp := &SFTP{
		log: log,
		config: &ssh.ClientConfig{
			Auth:            []ssh.AuthMethod{},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		},
		pool:                 xssh.NewConnectionPool(log),
		concurrency:          concurrency,
		maxClientConcurrency: maxClientConcurrency,
		maxPacketSize:        maxPacketSize,
		queue:                queue,
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

func (s *SFTP) GetConnection(address, user string) (*xssh.Connection, error) {
	return s.pool.GetConnection(address, user, s.config, s.concurrency, s.maxClientConcurrency, s.maxPacketSize)
}

func (s *SFTP) Get(uri *url.URL, w io.Writer) (int64, error) {
	if uri.Scheme != "sftp" {
		return 0, fmt.Errorf("invalid uri scheme %s for sftp", uri.Scheme)
	}
	user := ""
	if uri.User != nil {
		user = uri.User.Username()
	}
	conn, err := s.GetConnection(uri.Host, user)
	if err != nil {
		return 0, emperror.Wrapf(err, "unable to connect to %v with user %v", uri.Host, user)
	}

	written, err := conn.ReadFile(uri.Path, w)
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
	if uri.Scheme != "sftp" {
		return 0, fmt.Errorf("invalid uri scheme %s for sftp", uri.Scheme)
	}
	userInfo := uri.User
	conn, err := s.GetConnection(uri.Host, userInfo.Username())
	if err != nil {
		return 0, emperror.Wrapf(err, "unable to connect to %v with user %v", uri.Host, userInfo.Username())
	}
	daReader := r
	for _, queueEntry := range s.queue {
		daReader = queueEntry.StartReader(daReader)
	}
	start := time.Now()
	written, err := conn.WriteFile(uri.Path, daReader)
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
