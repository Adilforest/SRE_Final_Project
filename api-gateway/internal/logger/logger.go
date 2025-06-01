package logger
import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
}

type LogrusLogger struct {
	*logrus.Logger
}

func NewLogrusLoggerToFile(filepath string) (Logger, error) {
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	logger := logrus.New()
	logger.SetOutput(file)
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)
	return &LogrusLogger{logger}, nil
}

func (l *LogrusLogger) SetOutput(w io.Writer) {
	l.Logger.SetOutput(w)
}

func (l *LogrusLogger) SetLevel(level logrus.Level) {
	l.Logger.SetLevel(level)
}
