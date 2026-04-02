package goapi

import (
	"fmt"
	"strings"
	"time"
)

type Logger interface {
	Debug(format string, a ...any)
	Info(format string, a ...any)
	Warning(format string, a ...any)
	Error(format string, a ...any)
	Fatal(format string, a ...any)
}

type LoggerContext interface {
	SetContext(ctx *Context)
}

type DefaultLogger struct {
}

func (d *DefaultLogger) Debug(format string, a ...any) {
	format = addStrings(spanFill(ColorDebug("DEBUG"), 5, 10), "[",
		time.Now().Format("2006-01-02 15:04:05"), "] ", format)
	format = fmt.Sprintf(format, a...)
	format = strings.ReplaceAll(format, "\n", "\n"+spanFill("", 0, 10))
	fmt.Printf(addStrings(format, "\n"))
}

func (d *DefaultLogger) Info(format string, a ...any) {
	format = addStrings(spanFill(ColorInfo("INFO"), 4, 10), "[",
		time.Now().Format("2006-01-02 15:04:05"), "] ", format)
	format = fmt.Sprintf(format, a...)
	format = strings.ReplaceAll(format, "\n", "\n"+spanFill("", 0, 10))
	fmt.Printf(addStrings(format, "\n"))
}

func (d *DefaultLogger) Warning(format string, a ...any) {
	format = addStrings(spanFill(ColorWarning("WARNING"), 7, 10), "[",
		time.Now().Format("2006-01-02 15:04:05"), "] ", format)
	format = fmt.Sprintf(format, a...)
	format = strings.ReplaceAll(format, "\n", "\n"+spanFill("", 0, 10))
	fmt.Printf(addStrings(format, "\n"))
}

func (d *DefaultLogger) Error(format string, a ...any) {
	format = addStrings(spanFill(ColorError("ERROR"), 5, 10), "[",
		time.Now().Format("2006-01-02 15:04:05"), "] ", format)
	format = fmt.Sprintf(format, a...)
	format = strings.ReplaceAll(format, "\n", "\n"+spanFill("", 0, 10))
	fmt.Printf(addStrings(format, "\n"))
}

func (d *DefaultLogger) Fatal(format string, a ...any) {
	format = addStrings(spanFill(ColorFatal("FATAL"), 5, 10), "[",
		time.Now().Format("2006-01-02 15:04:05"), "] ", format)
	format = fmt.Sprintf(format, a...)
	format = strings.ReplaceAll(format, "\n", "\n"+spanFill("", 0, 10))
	fmt.Printf(addStrings(format, "\n"))
}

var Colorful = true
