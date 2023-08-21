package logger

import (
	"io"
	"log"
	"os"
)

var (
	logFlags                  = log.Ldate | log.Ltime
	LoggerWritter io.Writer   = os.Stderr
	WarningLogger *log.Logger = log.New(LoggerWritter, "WARNING: ", logFlags)
	InfoLogger    *log.Logger = log.New(LoggerWritter, "INFO: ", logFlags)
	ErrorLogger   *log.Logger = log.New(LoggerWritter, "ERROR: ", logFlags)
)

const logPath = "/var/log/eic"

func StartLogger() {
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		errMkDir := os.Mkdir(logPath, 0777)
		if errMkDir != nil {
			ErrorLogger.Println("Cannot create log's directory " + errMkDir.Error())
		}
	}

	file, err := os.OpenFile(logPath+"/app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		ErrorLogger.Println("Cannot save log file " + err.Error())
	} else {
		LoggerWritter = io.MultiWriter(os.Stdout, file)
	}

	InfoLogger = log.New(LoggerWritter, "INFO: ", logFlags)
	WarningLogger = log.New(LoggerWritter, "WARNING: ", logFlags)
	ErrorLogger = log.New(LoggerWritter, "ERROR: ", logFlags)
}
