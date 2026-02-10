## [<<](examples.md) 将gorm日志嵌入系统日志
### 系统Logger日志转gorm日志逻辑
~~~go
package gormLogger

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/goodluckxu-go/goapi/v2"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

func LoggerConvertGormLogger(log goapi.Logger) logger.Interface {
	return NewGormLogger(&GormWriter{log: log}, logger.Config{
		SlowThreshold: time.Second, // 慢 SQL 阈值
		LogLevel:      logger.Info, // Log level
		Colorful:      false,       // 禁用彩色打印
	})
}

type GormWriter struct {
	log goapi.Logger
}

func (w *GormWriter) Printf(format string, a ...interface{}) {
	if strings.HasPrefix(format, "GORM-ERROR") {
		w.log.Error(format, a...)
	} else {
		w.log.Debug(format, a...)
	}
}

func NewGormLogger(writer logger.Writer, config logger.Config) logger.Interface {
	var (
		infoStr       = "GORM-INFO %s\n"
		warnStr       = "GORM-WARNING %s\n"
		errStr        = "GORM-ERROR %s\n"
		traceStr      = "GORM-INFO %s\n[%.3fms] [rows:%v] %s"
		traceWarnStr1 = "GORM-WARNING %s %s\n[%.3fms] [rows:%v] %s"
		traceWarnStr2 = "GORM-WARNING %s %s\n[%.3fms] [rows:%v] %s"
		traceErrStr   = "GORM-ERROR %s %s\n[%.3fms] [rows:%v] %s"
	)
	if config.Colorful {
		infoStr = "GORM-INFO %s\n"
		warnStr = "GORM-WARNING %s\n"
		errStr = "GORM-ERROR %s\n"
		traceStr = "GORM-INFO %s\n" + colorGreen("[%.3fms]") + " " + colorCyan("[rows:%v]") + " %s"
		traceWarnStr1 = "GORM-WARNING %s " + colorYellow("%s") + "\n" + colorGreen("[%.3fms]") + " " + colorCyan("[rows:%v]") + " %s"
		traceWarnStr2 = "GORM-WARNING %s " + colorYellow("%s") + "\n" + colorYellow("[%.3fms]") + " " + colorCyan("[rows:%v]") + " %s"
		traceErrStr = "GORM-ERROR %s " + colorRed("%s") + "\n" + colorRed("[%.3fms]") + " " + colorCyan("[rows:%v]") + " %s"
	}
	return &GormLogger{
		writer,
		config,
		infoStr, warnStr, errStr,
		traceStr, traceErrStr, traceWarnStr1, traceWarnStr2,
	}
}

type GormLogger struct {
	logger.Writer
	logger.Config
	infoStr, warnStr, errStr                            string
	traceStr, traceErrStr, traceWarnStr1, traceWarnStr2 string
}

func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newlogger := *l
	newlogger.LogLevel = level
	return &newlogger
}

func (l GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Info {
		l.Printf(l.infoStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

func (l GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Warn {
		l.Printf(l.warnStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

func (l GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Error {
		l.Printf(l.errStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

func (l GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}
	elapsed := time.Since(begin)
	sql, rows := fc()
	if err == nil {
		if elapsed > l.SlowThreshold && l.SlowThreshold != 0 {
			slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
			if rows == -1 {
				l.Printf(l.traceWarnStr2, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
			} else {
				l.Printf(l.traceWarnStr2, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
			}
		} else {
			if rows == -1 {
				l.Printf(l.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, "-", sql)
			} else {
				l.Printf(l.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
			}
		}
	} else {
		if errors.Is(err, logger.ErrRecordNotFound) {
			if l.IgnoreRecordNotFoundError {
				if rows == -1 {
					l.Printf(l.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, "-", sql)
				} else {
					l.Printf(l.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
				}
			} else {
				if rows == -1 {
					l.Printf(l.traceWarnStr1, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
				} else {
					l.Printf(l.traceWarnStr1, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
				}
			}
		} else {
			if rows == -1 {
				l.Printf(l.traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
			} else {
				l.Printf(l.traceErrStr, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
			}
		}
	}
}

func (l GormLogger) ParamsFilter(ctx context.Context, sql string, params ...interface{}) (string, []interface{}) {
	if l.Config.ParameterizedQueries {
		return sql, nil
	}
	return sql, params
}

var colorGreen = color.New(color.FgGreen).SprintFunc()
var colorCyan = color.New(color.FgCyan).SprintFunc()
var colorYellow = color.New(color.FgHiYellow).SprintFunc()
var colorRed = color.New(color.FgRed).SprintFunc()
~~~
### 全局使用日志
~~~go
api := goapi.GoAPI(true)
// 全局模式
db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{
	Logger: LoggerConvertGormLogger(api.Logger()),
})
~~~
### 每次请求当一次会话使用日志
~~~go
func (*Index)Index(ctx *goapi.Context, input struct{
	router goapi.Router `paths:"/index" methods:"GET"`
}) {
	// 新建会话模式 
	tx := db.Session(&Session{Logger: LoggerConvertGormLogger(ctx.Logger())})
	tx.First(&user)
	tx.Model(&user).Update("Age", 18)
}
~~~