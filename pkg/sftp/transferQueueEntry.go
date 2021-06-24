package sftp

import "io"

type TransferQueueEntry interface {
	StartReader(reader io.Reader) io.Reader
	StartWriter(writer io.Writer) io.Writer
}
