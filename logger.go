package goapi

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
)

type Logger interface {
	Debug(format string, a ...any)
	Info(format string, a ...any)
	Warning(format string, a ...any)
	Error(format string, a ...any)
	Fatal(format string, a ...any)
}

type LoggerRequestParam interface {
	SetRequestParam(childPath, requestID string)
}

type defaultLogger struct {
}

func (d *defaultLogger) Debug(format string, a ...any) {
	format = fmt.Sprintf(format, a...)
	format = strings.ReplaceAll(format, "\n", "\n"+spanFill("", 0, 10))
	fmt.Printf("%v", spanFill(colorDebug("DEBUG"), 5, 10)+"["+timeFormat(time.Now())+"] "+format+"\n")
}

func (d *defaultLogger) Info(format string, a ...any) {
	format = fmt.Sprintf(format, a...)
	format = strings.ReplaceAll(format, "\n", "\n"+spanFill("", 0, 10))
	fmt.Printf("%v", spanFill(colorInfo("INFO"), 4, 10)+"["+timeFormat(time.Now())+"] "+format+"\n")
}

func (d *defaultLogger) Warning(format string, a ...any) {
	format = fmt.Sprintf(format, a...)
	format = strings.ReplaceAll(format, "\n", "\n"+spanFill("", 0, 10))
	fmt.Printf("%v", spanFill(colorWarning("WARNING"), 7, 10)+"["+timeFormat(time.Now())+"] "+format+"\n")
}

func (d *defaultLogger) Error(format string, a ...any) {
	format = fmt.Sprintf(format, a...)
	format = strings.ReplaceAll(format, "\n", "\n"+spanFill("", 0, 10))
	fmt.Printf("%v", spanFill(colorError("ERROR"), 5, 10)+"["+timeFormat(time.Now())+"] "+format+"\n")
}

func (d *defaultLogger) Fatal(format string, a ...any) {
	format = fmt.Sprintf(format, a...)
	format = strings.ReplaceAll(format, "\n", "\n"+spanFill("", 0, 10))
	fmt.Printf("%v", spanFill(colorFatal("FATAL"), 5, 10)+"["+timeFormat(time.Now())+"] "+format+"\n")
}

type levelHandleLogger struct {
	log Logger
}

func (d *levelHandleLogger) Debug(format string, a ...any) {
	if d.log == nil {
		return
	}
	if logLevel&LogDebug == 0 {
		return
	}
	d.log.Debug(format, a...)
}

func (d *levelHandleLogger) Info(format string, a ...any) {
	if d.log == nil {
		return
	}
	if logLevel&LogInfo == 0 {
		return
	}
	d.log.Info(format, a...)
}

func (d *levelHandleLogger) Warning(format string, a ...any) {
	if d.log == nil {
		return
	}
	if logLevel&LogWarning == 0 {
		return
	}
	d.log.Warning(format, a...)
}

func (d *levelHandleLogger) Error(format string, a ...any) {
	if d.log == nil {
		return
	}
	if logLevel&LogError == 0 {
		return
	}
	d.log.Error(format, a...)
}

func (d *levelHandleLogger) Fatal(format string, a ...any) {
	if d.log == nil {
		return
	}
	if logLevel&LogFail == 0 {
		return
	}
	d.log.Fatal(format, a...)
}

var colorInfo = color.New(color.FgGreen).SprintFunc()
var colorDebug = color.New(color.FgCyan).SprintFunc()
var colorWarning = color.New(color.FgHiYellow).SprintFunc()
var colorError = color.New(color.FgRed).SprintFunc()
var colorFatal = color.New(color.BgRed, color.FgWhite).SprintFunc()

func SetLogLevel(level LogLevel) {
	logLevel = level
}
