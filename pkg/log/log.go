package log

import "github.com/sirupsen/logrus"

// InitLogging sets up logging using predetermined defaults
func InitLogging() {
	logrus.SetLevel(logrus.DebugLevel)

	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
}
