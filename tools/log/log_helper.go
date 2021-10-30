package log

import (
	"fmt"
	"io"
	"log"
	"log/syslog"
	"os"
	"path/filepath"
)

const (
	flag = log.Ldate | log.Ltime | log.Lshortfile
	preDebug = "[DEBUG]"
	preInfo = "[INFO]"
	preWarning = "[WARNING]"
	preError = "[ERROR]"
)

var (
	logFile io.Writer
	debugLogger *log.Logger
	infoLogger *log.Logger
	warningLogger *log.Logger
	errorLogger *log.Logger
	defaultLogFile = "./data/log/glue.log"
)

func setUp() {
	var err error
	logFile, err = os.OpenFile(defaultLogFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		if err != nil {
			log.Fatalf("create log file failed, error: %+v", err)
		}
	}
	debugLogger = log.New(logFile, preDebug, flag)
	infoLogger = log.New(logFile, preInfo, flag)
	warningLogger = log.New(logFile, preWarning, flag)
	errorLogger = log.New(logFile, preError, flag)
}

func Debugf(format string, v ...interface{}) {
	if debugLogger == nil {
		setUp()
	}
	debugLogger.Printf(format, v...)
}

func Infof(format string, v ...interface{}) {
	if debugLogger == nil {
		setUp()
	}
	infoLogger.Printf(format, v...)
}

func Warningf(format string, v ...interface{}) {
	if debugLogger == nil {
		setUp()
	}
	warningLogger.Printf(format, v...)
}

func ErrorF(format string, v ...interface{}) {
	if debugLogger == nil {
		setUp()
	}
	errorLogger.Printf(format, v...)
}

func SetOutputPath(path string) {
	var err error
	logFile, err = os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		log.Fatalf("create log file failed, error: %+v", err)
	}
	debugLogger.SetOutput(logFile)
	infoLogger.SetOutput(logFile)
	warningLogger.SetOutput(logFile)
	errorLogger.SetOutput(logFile)
}

func WriteLog() {
	programName := filepath.Base(os.Args[0])
	sysLog, err := syslog.New(syslog.LOG_INFO|syslog.LOG_LOCAL7, programName)
	if err != nil {
		log.Fatal(err)
	}else {
		log.SetOutput(sysLog)
	}
	log.Println("LOG_INFO + LOG_LOCAL7: Logging in Go!")

	sysLog, err = syslog.New(syslog.LOG_MAIL, "Some program!")
	if err != nil {
		log.Fatal(err)
	}else {
		log.SetOutput(sysLog)
	}
	log.Println("LOG_MAIL: Logging in Go!")
	fmt.Println("Will you see this?")
}
