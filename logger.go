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

type defaultLogger struct {
}

func (d *defaultLogger) Debug(format string, a ...any) {
	format = fmt.Sprintf(format, a...)
	format = strings.ReplaceAll(format, "\n", "\n"+spanFill("", 0, 10))
	fmt.Printf("%v", spanFill(ColorDebug("DEBUG"), 5, 10)+"["+timeFormat(time.Now())+"] "+format+"\n")
}

func (d *defaultLogger) Info(format string, a ...any) {
	format = fmt.Sprintf(format, a...)
	format = strings.ReplaceAll(format, "\n", "\n"+spanFill("", 0, 10))
	fmt.Printf("%v", spanFill(ColorInfo("INFO"), 4, 10)+"["+timeFormat(time.Now())+"] "+format+"\n")
}

func (d *defaultLogger) Warning(format string, a ...any) {
	format = fmt.Sprintf(format, a...)
	format = strings.ReplaceAll(format, "\n", "\n"+spanFill("", 0, 10))
	fmt.Printf("%v", spanFill(ColorWarning("WARNING"), 7, 10)+"["+timeFormat(time.Now())+"] "+format+"\n")
}

func (d *defaultLogger) Error(format string, a ...any) {
	format = fmt.Sprintf(format, a...)
	format = strings.ReplaceAll(format, "\n", "\n"+spanFill("", 0, 10))
	fmt.Printf("%v", spanFill(ColorError("ERROR"), 5, 10)+"["+timeFormat(time.Now())+"] "+format+"\n")
}

func (d *defaultLogger) Fatal(format string, a ...any) {
	format = fmt.Sprintf(format, a...)
	format = strings.ReplaceAll(format, "\n", "\n"+spanFill("", 0, 10))
	fmt.Printf("%v", spanFill(ColorFatal("FATAL"), 5, 10)+"["+timeFormat(time.Now())+"] "+format+"\n")
}

var Colorful = true
