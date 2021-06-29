package stream

import (
	"context"
	"github.com/machinebox/progress"
	"io"
	"time"
)

type callbackFunc func(remaining time.Duration, percent float64, estimated time.Time, complete bool)

type ProgressReaderWriter struct {
	filesize int64
	interval time.Duration
	callback callbackFunc
}

func NewProgressReaderWriter(filesize int64, interval time.Duration, callback callbackFunc) *ProgressReaderWriter {
	pm := &ProgressReaderWriter{
		filesize: filesize,
		interval: interval,
		callback: callback,
	}
	return pm
}

func (pr *ProgressReaderWriter) StartReader(reader io.Reader) io.Reader {
	r2 := progress.NewReader(reader)
	go func() {
		ctx := context.Background()
		progressChan := progress.NewTicker(ctx, r2, pr.filesize, pr.interval)
		for p := range progressChan {
			pr.callback(p.Remaining(), p.Percent(), p.Estimated(), p.Complete())
		}
	}()
	return r2
}

func (pr *ProgressReaderWriter) StartWriter(writer io.Writer) io.Writer {
	w2 := progress.NewWriter(writer)
	go func() {
		ctx := context.Background()
		progressChan := progress.NewTicker(ctx, w2, pr.filesize, pr.interval)
		for p := range progressChan {
			pr.callback(p.Remaining(), p.Percent(), p.Estimated(), p.Complete())
			if p.Complete() {
				break
			}
		}
	}()
	return w2
}
