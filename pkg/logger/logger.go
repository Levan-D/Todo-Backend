package logger

import (
	"fmt"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"os"
)

func init() {
	_ = godotenv.Load(".env")

	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)

	if os.Getenv("LOG") == "debug" {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
}

func SetFileSaveData(fileName string, folderPath string) {
	if os.Getenv("MODE") != "development" {
		path := fmt.Sprintf("%s/%s.log", folderPath, fileName)
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			log.Fatal(err)
		}
		log.SetOutput(f)
	}
}

func Debug(fields log.Fields, data ...interface{}) {
	log.WithFields(fields).Debug(data...)
}

func Info(fields log.Fields, data ...interface{}) {
	log.WithFields(fields).Info(data...)
}

func Warn(fields log.Fields, data ...interface{}) {
	log.WithFields(fields).Warn(data...)
}

func Error(fields log.Fields, data ...interface{}) {
	log.WithFields(fields).Error(data...)
}
