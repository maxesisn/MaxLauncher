package logger

import (
	"github.com/withmandala/go-log"
	"os"
)

func Logger() *log.Logger {
	localLogger := log.New(os.Stdout).WithColor()
	return localLogger
}
