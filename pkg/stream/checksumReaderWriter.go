package stream

import (
	logger "github.com/op/go-logging"
	"hash"
	"io"
)

type ChecksumReaderWriter struct {
	mac    hash.Hash
	logger *logger.Logger
}

func NewChecksumReaderWriter(mac hash.Hash, logger *logger.Logger) *ChecksumReaderWriter {
	enc := &ChecksumReaderWriter{
		mac:    mac,
		logger: logger,
	}
	return enc
}

func (cr *ChecksumReaderWriter) StartReader(reader io.Reader) io.Reader {
	cr.mac.Reset()
	pr, pw := io.Pipe()
	tr := io.TeeReader(reader, pw)
	go func() {
		defer pw.Close()
		if _, err := io.Copy(cr.mac, pr); err != nil {
			cr.logger.Errorf("cannot read data: %v", err)
		}
	}()
	return tr
}

func (cr *ChecksumReaderWriter) StartWriter(writer io.Writer) io.Writer {
	return io.MultiWriter(writer, cr.mac)
}
