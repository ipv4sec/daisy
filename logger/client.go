package logger

import (
	"fmt"
	"io"
	"os"
	"time"
)

func Info(msg ...interface{}) {
	log(os.Stdout, "INFO", msg...)
}
func Error(msg ...interface{}) {
	log(os.Stderr, "ERROR", msg...)
}

func log(out io.Writer, logType string, msg ...interface{}) {
	var args []interface{}
	args = append(args, "["+time.Now().Format("2006-01-02 15:04:05")+"]")
	args = append(args, "["+logType+"]")
	for _, v := range msg {
		args = append(args, v)
	}
	fmt.Fprintln(out, args...)
}