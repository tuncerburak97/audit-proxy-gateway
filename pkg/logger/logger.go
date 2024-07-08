package logger

import (
	"audit-proxy-gateway/internal/config"
	"github.com/sirupsen/logrus"
	"strings"
)

var log = logrus.New()

func InitLogger() {
	cfg := config.GetConfig()
	level, err := logrus.ParseLevel(strings.ToLower(cfg.Application.Log.Level))
	if err != nil {
		log.Fatalf("Invalid log level: %s", cfg.Application.Log.Level)
	}
	log.SetLevel(level)
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
}

func GetLogger() *logrus.Logger {
	return log
}
