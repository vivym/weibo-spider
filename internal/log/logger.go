package log

import (
	"os"

	"github.com/sirupsen/logrus"
	logrusAdapter "logur.dev/adapter/logrus"
	"logur.dev/logur"
)

func NewLogger(config Config) logur.LoggerFacade {
	logger := logrus.New()

	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&logrus.TextFormatter{
		DisableColors:             config.NoColor,
		EnvironmentOverrideColors: true,
	})

	switch config.Format {
	case "logfmt":
		// Already the default
	case "json":
		logger.SetFormatter(&logrus.JSONFormatter{})
	}

	if level, err := logrus.ParseLevel(config.Level); err == nil {
		logger.SetLevel(level)
	}

	return logrusAdapter.New(logger)
}
