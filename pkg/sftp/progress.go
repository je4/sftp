package sftp

import (
	"context"
	"github.com/machinebox/progress"
	"io"
	"time"
)

type callbackFunc func(remaining time.Duration, percent float64, estimated time.Time, complete bool)

type Progress struct {
	filesize int64
	interval time.Duration
	callback callbackFunc
}

func NewProgress(filesize int64, interval time.Duration, callback callbackFunc) *Progress {
	pm := &Progress{
		filesize: filesize,
		interval: interval,
		callback: callback,
	}
	return pm
}

func (pm *Progress) StartReader(reader io.Reader) io.Reader {
	r2 := progress.NewReader(reader)
	go func() {
		ctx := context.Background()
		progressChan := progress.NewTicker(ctx, r2, pm.filesize, pm.interval)
		for p := range progressChan {
			pm.callback(p.Remaining(), p.Percent(), p.Estimated(), p.Complete())
		}
	}()
	return r2
}

func (pm *Progress) StartWriter(writer io.Writer) io.Writer {
	w2 := progress.NewWriter(writer)
	go func() {
		ctx := context.Background()
		progressChan := progress.NewTicker(ctx, w2, pm.filesize, pm.interval)
		for p := range progressChan {
			pm.callback(p.Remaining(), p.Percent(), p.Estimated(), p.Complete())
		}
	}()
	return w2
}
