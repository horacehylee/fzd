package main

import (
	"os"

	"github.com/sirupsen/logrus"
)

func newLogger() *logrus.Logger {
	log := logrus.New()
	log.Formatter = new(logrus.TextFormatter)
	log.Formatter.(*logrus.TextFormatter).DisableColors = true
	log.Formatter.(*logrus.TextFormatter).FullTimestamp = true
	log.Level = logrus.InfoLevel

	log.Out = os.Stdout
	return log
}
