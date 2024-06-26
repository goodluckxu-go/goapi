package goapi

import (
	"fmt"
	"github.com/fatih/color"
	"time"
)

type Logger interface {
	Debug(format string, a ...any)
	Info(format string, a ...any)
	Warning(format string, a ...any)
	Error(format string, a ...any)
	Fatal(format string, a ...any)
}

type defaultLogger struct {
}

func (d *defaultLogger) Debug(format string, a ...any) {
	fmt.Printf(spanFill(colorDebug("DEBUG"), 5, 10)+" ["+timeFormat(time.Now())+"] "+format+"\n", a...)
}

func (d *defaultLogger) Info(format string, a ...any) {
	fmt.Printf(spanFill(colorInfo("INFO"), 4, 10)+" ["+timeFormat(time.Now())+"] "+format+"\n", a...)
}

func (d *defaultLogger) Warning(format string, a ...any) {
	fmt.Printf(spanFill(colorWarning("WARNING"), 7, 10)+" ["+timeFormat(time.Now())+"] "+format+"\n", a...)
}

func (d *defaultLogger) Error(format string, a ...any) {
	fmt.Printf(spanFill(colorError("ERROR"), 5, 10)+" ["+timeFormat(time.Now())+"] "+format+"\n", a...)
}

func (d *defaultLogger) Fatal(format string, a ...any) {
	fmt.Printf(spanFill(colorFatal("FATAL"), 5, 10)+" ["+timeFormat(time.Now())+"] "+format+"\n", a...)
}

var colorInfo = color.New(color.FgGreen).SprintFunc()
var colorDebug = color.New(color.FgCyan).SprintFunc()
var colorWarning = color.New(color.FgHiYellow).SprintFunc()
var colorError = color.New(color.FgRed).SprintFunc()
var colorFatal = color.New(color.BgRed, color.FgWhite).SprintFunc()
