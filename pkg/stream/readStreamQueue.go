package stream

import "io"

type ReadStreamQueue struct {
	queue []ReadQueueEntry
}

func NewReadStreamQueue(entries ...ReadQueueEntry) (*ReadStreamQueue, error) {
	rsq := &ReadStreamQueue{queue: entries}
	return rsq, nil
}

func (rsq ReadStreamQueue) Append(entry ReadQueueEntry) {
	rsq.queue = append(rsq.queue, entry)
}

func (rsq ReadStreamQueue) StartReader(reader io.Reader) io.Reader {
	for _, e := range rsq.queue {
		reader = e.StartReader(reader)
	}
	return reader
}
