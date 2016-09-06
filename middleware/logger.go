package middleware

import (
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/setting"
)

func InitLogger() {
	switch strings.ToLower(setting.LogLevel) {
	case "panic":
		log.SetLevel(log.PanicLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	default:
		if setting.RunMode == "dev" {
			log.SetLevel(log.DebugLevel)
		} else {
			log.SetLevel(log.ErrorLevel)
		}
	}

	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{DisableColors: true})
	file, err := os.OpenFile(setting.LogPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0640)
	if err != nil {
		log.Errorf("Failed to init logger, can not open log file: %v", setting.LogPath)
		return
	}
	log.SetOutput(file)
}

func logger() macaron.Handler {
	return func(ctx *macaron.Context) {
		//trace log in dev DebugLevel
		log.WithFields(log.Fields{
			"Method": ctx.Req.Method,
			"URL":    ctx.Req.RequestURI,
		}).Info(ctx.Req.Header)
	}
}
