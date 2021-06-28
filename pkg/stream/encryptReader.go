package stream

import (
	"crypto/cipher"
	"github.com/blend/go-sdk/crypto"
	logger "github.com/op/go-logging"
	"hash"
	"io"
)

const (
	BUFSIZE = 32 * 1024
)

type EncryptReader struct {
	stream cipher.Stream
	block  cipher.Block
	mac    hash.Hash
	iv     []byte
	logger *logger.Logger
}

func NewEncryptReader(block cipher.Block, stream cipher.Stream, mac hash.Hash, iv []byte, logger *logger.Logger) *EncryptReader {
	enc := &EncryptReader{
		block:  block,
		stream: stream,
		mac:    mac,
		iv:     iv,
		logger: logger,
	}
	return enc
}

func (er *EncryptReader) StartReader(reader io.Reader) io.Reader {
	//tr := io.TeeReader(reader, reader)
	enc := &crypto.StreamEncrypter{
		Source: reader,
		Block:  er.block,
		Stream: er.stream,
		Mac:    er.mac,
		IV:     er.iv,
	}

	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		var rbuf = make([]byte, BUFSIZE)
		for {
			n, err := enc.Read(rbuf)
			if err != nil {
				break
			}
			n, err = pw.Write(rbuf[:n])
			if err != nil {
				break
			}
			if n < BUFSIZE {
				break
			}
		}
	}()

	return pr
}
