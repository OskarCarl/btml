package logging

import (
	"log"
	"os"
	"strconv"
	"time"
)

var Logger *MonotonicLogWriter

func init() {
	Logger = &MonotonicLogWriter{}
}

type MonotonicLogWriter struct {
	prefix,
	buf []byte
	prefixLen int
}

func (l *MonotonicLogWriter) Use() {
	log.Default().SetOutput(Logger)
	log.Default().SetFlags(0)
}

func (l *MonotonicLogWriter) SetPrefix(prefix string) {
	l.prefix = []byte(prefix)
	l.prefixLen = len(l.prefix) + 1
	l.buf = make([]byte, 0, 128)
	l.buf = append(l.buf, l.prefix...)
	l.buf = append(l.buf, 0x20)
}

func (l *MonotonicLogWriter) Write(p []byte) (int, error) {
	t := strconv.FormatInt(time.Now().UnixMicro(), 10)
	l.buf = l.buf[:l.prefixLen]
	l.buf = append(l.buf, t...)
	l.buf = append(l.buf, 0x20)
	l.buf = append(l.buf, p...)
	return os.Stdout.Write(l.buf)
}
