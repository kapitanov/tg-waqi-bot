package log

import (
	"fmt"
	"log"
)

type Logger interface {
	Printf(format string, v ...interface{})
	Fatalf(format string, v ...interface{})
}

func New(prefix string) Logger {
	return &logger{prefix}
}

type logger struct {
	prefix string
}

func (l *logger) Printf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	log.Printf("%s: %s", l.prefix, msg)
}

func (l *logger) Fatalf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	log.Fatalf("%s: %s", l.prefix, msg)
}
