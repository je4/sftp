package sftp

import (
	logger "github.com/op/go-logging"
	"hash"
	"io"
)

type Checksum struct {
	mac    hash.Hash
	logger *logger.Logger
}

func NewChecksum(mac hash.Hash, logger *logger.Logger) *Checksum {
	enc := &Checksum{
		mac:    mac,
		logger: logger,
	}
	return enc
}

func (c *Checksum) StartReader(reader io.Reader) io.Reader {
	pr, pw := io.Pipe()
	tr := io.TeeReader(reader, pw)
	go func() {
		defer pw.Close()
		if _, err := io.Copy(c.mac, pr); err != nil {
			c.logger.Errorf("cannot read data: %v", err)
		}
	}()
	return tr
}

func (c *Checksum) StartWriter(writer io.Writer) io.Writer {
	return io.MultiWriter(writer, c.mac)
}
