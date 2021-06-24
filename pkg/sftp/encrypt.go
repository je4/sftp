package sftp

import (
	"crypto/cipher"
	"github.com/blend/go-sdk/crypto"
	logger "github.com/op/go-logging"
	"github.com/tidwall/transform"
	"hash"
	"io"
)

type Encrypt struct {
	stream cipher.Stream
	block  cipher.Block
	mac    hash.Hash
	iv     []byte
	logger *logger.Logger
}

func NewEncrypt(block cipher.Block, stream cipher.Stream, mac hash.Hash, iv []byte, logger *logger.Logger) *Encrypt {
	enc := &Encrypt{
		block:  block,
		stream: stream,
		mac:    mac,
		iv:     iv,
		logger: logger,
	}
	return enc
}

func (e *Encrypt) StartReader(reader io.Reader) io.Reader {
	//tr := io.TeeReader(reader, reader)
	enc := &crypto.StreamEncrypter{
		Source: reader,
		Block:  e.block,
		Stream: e.stream,
		Mac:    e.mac,
		IV:     e.iv,
	}

	var rbuf = make([]byte, 4096)
	r := transform.NewTransformer(func() ([]byte, error) {
		var err error
		n, err := enc.Read(rbuf)
		if err != nil {
			return nil, err
		}
		return rbuf[:n], nil
	})
	return r
}

func (e *Encrypt) StartWriter(writer io.Writer) io.Writer {
	panic("func (e *Encrypt) StartWriter(writer io.Writer) io.Writer not implemented")
}
