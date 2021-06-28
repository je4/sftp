package stream

import "io"

type ReadQueueEntry interface {
	StartReader(reader io.Reader) io.Reader
}

type WriteQueueEntry interface {
	StartWriter(writer io.Writer) io.Writer
}

type RWQueueEntry interface {
	ReadQueueEntry
	WriteQueueEntry
}
