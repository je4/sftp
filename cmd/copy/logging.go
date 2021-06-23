package main

import (
	"github.com/op/go-logging"
	"io"
	"os"
)

func CreateLogger(module string, logfile string, w *io.PipeWriter, loglevel string, logformat string) (log *logging.Logger, lf *os.File) {
	log = logging.MustGetLogger(module)
	var err error
	if logfile != "" {
		lf, err = os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Errorf("Cannot open logfile %v: %v", logfile, err)
		}
		//defer lf.Close()

	} else {
		lf = os.Stderr
	}
	var w2 io.Writer
	if w != nil {
		w2 = io.MultiWriter(w, lf)
	} else {
		w2 = lf
	}
	backend := logging.NewLogBackend(w2, "", 0)
	backendLeveled := logging.AddModuleLevel(backend)
	backendLeveled.SetLevel(logging.GetLevel(loglevel), "")

	logging.SetFormatter(logging.MustStringFormatter(logformat))
	logging.SetBackend(backendLeveled)

	return
}
