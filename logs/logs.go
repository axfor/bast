//Copyright 2018 The axx Authors. All rights reserved.

package logs

import (
	"database/sql/driver"
	"fmt"
	"log"
	"os"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

var (
	logger                       *Log
	gromDebugLogger              = log.New(os.Stdout, "\r\n", 0)
	gromSQLRegexp                = regexp.MustCompile(`\?`)
	gromNumericPlaceHolderRegexp = regexp.MustCompile(`\$\d+`)
)

//Conf log config
type Conf struct {
	OutPath    string `json:"outPath"`
	Level      string `json:"level"`
	MaxSize    int    `json:"maxSize"`
	MaxBackups int    `json:"maxBackups"`
	MaxAge     int    `json:"maxAge"`
	Debug      bool   `json:"debug"`
	LogSelect  bool   `json:"logSelect"`
	Stdout     bool   `json:"-"`
}

//Log log
type Log struct {
	zap.Logger
	logConf *Conf
}

//GormLog Gorm loger
type GormLog struct {
}

//Init init log
func Init(conf *Conf) *Log {
	if logger == nil {
		logger = newLog(conf)
	}
	return logger
}

func newLog(conf *Conf) *Log {
	if conf == nil {
		conf = &Conf{
			OutPath:    "./logs/logs.log",
			Level:      "info",
			MaxSize:    10,
			MaxBackups: 5,
			MaxAge:     28,
			Debug:      false,
		}
	} else {
		if conf.OutPath == "" {
			conf.Stdout = true
		}
	}
	if conf.MaxSize <= 0 {
		conf.MaxSize = 10
	}
	if conf.MaxBackups <= 0 {
		conf.MaxSize = 5
	}
	if conf.MaxAge <= 0 {
		conf.MaxAge = 28
	}

	l := logLevel(conf.Level)
	var w zapcore.WriteSyncer
	var core zapcore.Core
	if !conf.Stdout {
		encoderConfig := zap.NewProductionEncoderConfig()
		//encoderConfig.LineEnding = zapcore.DefaultLineEnding
		encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05"))
		}
		w = zapcore.AddSync(&lumberjack.Logger{
			Filename:   conf.OutPath,
			MaxSize:    conf.MaxSize, // megabytes
			MaxBackups: conf.MaxBackups,
			MaxAge:     conf.MaxAge, // days
		})
		core = zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			w,
			l,
		)
	} else {
		encoderConfig := zap.NewDevelopmentEncoderConfig()
		encoderConfig.LineEnding = zapcore.DefaultLineEnding
		encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05"))
		}

		//jsonDebugging := zapcore.AddSync(ioutil.Discard)
		//jsonErrors := zapcore.AddSync(ioutil.Discard)
		consoleDebugging := zapcore.Lock(os.Stdout)
		consoleErrors := zapcore.Lock(os.Stderr)

		//jsonEncoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
		consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

		core = zapcore.NewTee(
			//zapcore.NewCore(jsonEncoder, jsonErrors, highPriority),
			zapcore.NewCore(consoleEncoder, consoleErrors, zapcore.FatalLevel),
			//zapcore.NewCore(jsonEncoder, jsonDebugging, lowPriority),
			zapcore.NewCore(consoleEncoder, consoleDebugging, zapcore.DebugLevel),
		)

		// w, _, _ = zap.Open("stdout")
		// core = zapcore.NewCore(
		// 	zapcore.NewConsoleEncoder(encoderConfig),
		// 	w,
		// 	l,
		// )
	}
	return &Log{Logger: *zap.New(core), logConf: conf}
}

// Info logs a message at InfoLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (l *Log) Info(msg string, fields ...zap.Field) {
	if msg != "" {
		l.Logger.Info(msg, fields...)
	}
}

// Info logs a message at InfoLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Info(msg string, fields ...zap.Field) {
	if logger != nil && msg != "" {
		logger.Info(msg, fields...)
	}
}

// Warn logs a message at WarnLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (l *Log) Warn(msg string, fields ...zap.Field) {
	if msg != "" {
		l.Logger.Warn(msg, fields...)
	}
}

// Warn logs a message at WarnLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Warn(msg string, fields ...zap.Field) {
	if logger != nil && msg != "" {
		logger.Warn(msg, fields...)
	}
}

// Debug logs a message at DebugLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (l *Log) Debug(msg string, fields ...zap.Field) {
	if msg != "" {
		c := caller()
		if fields != nil {
			fields = append(fields, c)
			l.Logger.Debug(msg, fields...)
		} else {
			l.Logger.Debug(msg, c)
		}
	}
}

// Debug logs a message at DebugLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Debug(msg string, fields ...zap.Field) {
	if logger != nil && msg != "" {
		logger.Debug(msg, fields...)
	}
}

// Error logs a message at ErrorLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (l *Log) Error(msg string, fields ...zap.Field) {
	if msg != "" {
		c := caller()
		if fields != nil {
			fields = append(fields, c)
			l.Logger.Error(msg, fields...)
		} else {
			l.Logger.Error(msg, c)
		}
	}
}

// Error logs a message at ErrorLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Error(msg string, fields ...zap.Field) {
	if logger != nil && msg != "" {
		logger.Error(msg, fields...)
	}
}

// ErrorWithCaller logs a message at ErrorLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (l *Log) ErrorWithCaller(msg string, caller zap.Field, fields ...zap.Field) {
	if msg != "" {
		if fields != nil {
			fields = append(fields, caller)
			l.Logger.Error(msg, fields...)
		} else {
			l.Logger.Error(msg, caller)
		}
	}
}

// ErrorWithCaller logs a message at ErrorLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func ErrorWithCaller(msg string, caller zap.Field, fields ...zap.Field) {
	if logger != nil && msg != "" {
		logger.ErrorWithCaller(msg, caller, fields...)
	}
}

// Errors logs a message at ErrorLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (l *Log) Errors(msg string, err error) {
	if msg != "" {
		c := caller()
		if err != nil {
			l.Logger.Error(msg, c, zap.Error(err))
		} else {
			l.Logger.Error(msg, c)
		}
	}
}

// Errors logs a message at ErrorLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Errors(msg string, err error) {
	if logger != nil && msg != "" {
		logger.Errors(msg, err)
	}
}

// DPanic logs a message at DPanicLevel. The message includes any fields
// passed at the log site, as well as any fields accumulated on the logger.
//
// If the logger is in development mode, it then panics (DPanic means
// "development panic"). This is useful for catching errors that are
// recoverable, but shouldn't ever happen.
func (l *Log) DPanic(msg string, fields ...zap.Field) {
	if msg != "" {
		c := caller()
		if fields != nil {
			fields = append(fields, c)
			l.Logger.DPanic(msg, fields...)
		} else {
			l.Logger.DPanic(msg, c)
		}
	}
}

// DPanic logs a message at DPanicLevel. The message includes any fields
// passed at the log site, as well as any fields accumulated on the logger.
//
// If the logger is in development mode, it then panics (DPanic means
// "development panic"). This is useful for catching errors that are
// recoverable, but shouldn't ever happen.
func DPanic(msg string, fields ...zap.Field) {
	if logger != nil && msg != "" {
		logger.DPanic(msg, fields...)
	}
}

// Panic logs a message at PanicLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then panics, even if logging at PanicLevel is disabled.
func (l *Log) Panic(msg string, fields ...zap.Field) {
	if msg != "" {
		c := caller()
		if fields != nil {
			fields = append(fields, c)
			l.Logger.Panic(msg, fields...)
		} else {
			l.Logger.Panic(msg, c)
		}
	}
}

// Panic logs a message at PanicLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then panics, even if logging at PanicLevel is disabled.
func Panic(msg string, fields ...zap.Field) {
	if logger != nil && msg != "" {
		logger.Panic(msg, fields...)
	}
}

// Sync calls the underlying Core's Sync method, flushing any buffered log
// entries. Applications should take care to call Sync before exiting.
func (l *Log) Sync() {
	l.Logger.Sync()
}

// Sync calls the underlying Core's Sync method, flushing any buffered log
// entries. Applications should take care to call Sync before exiting.
func Sync() {
	if logger != nil {
		logger.Sync()
	}
}

//Default rew logger object
func Default() *Log {
	return logger
}

//New rew logger object
func New(conf *Conf) *Log {
	return newLog(conf)
}

//Gorm Gorm logger object
func Gorm() *GormLog {
	return &GormLog{}
}

func caller() zap.Field {
	return zap.String("caller", callerWithIndex(0))
}

func callerWithIndex(skip int) string {
	const callerSkipOffset = 3
	return zapcore.NewEntryCaller(runtime.Caller(skip + callerSkipOffset)).TrimmedPath()
}

//Print Gorm log
func (*GormLog) Print(v ...interface{}) {
	if logger == nil {
		return
	}
	if logger.logConf.Stdout {
		msg := gromLogFormatterDebug(v...)
		if msg != nil {
			gromDebugLogger.Println(msg...)
		}
	} else {
		msg, _ := gromLogFormatter(v...)
		if msg != nil {
			source, ok := v[1].(string)
			if ok && source != "" {
				msg = append(msg, zap.String("caller", source))
			}
			Info("gorm", msg...)
		}
	}
}

/*********************************
* grom
*********************************/

//grom logFormatter of debug
var gromLogFormatterDebug = func(values ...interface{}) (messages []interface{}) {
	if len(values) > 1 {
		var level = values[0]
		if level == "sql" && !logger.logConf.LogSelect {
			s, ok := values[3].(string)
			if ok && s != "" && strings.Index(s, "SELECT") >= 0 {
				return
			}
		}

		var (
			sql             string
			formattedValues []string
			currentTime     = "\n\033[33m[" + time.Now().Format("2006-01-02 15:04:05") + "]\033[0m"
			source          = fmt.Sprintf("\033[35m(%v)\033[0m", values[1])
		)

		messages = []interface{}{source, currentTime}

		if level == "sql" {
			// duration
			messages = append(messages, fmt.Sprintf(" \033[36;1m[%.2fms]\033[0m ", float64(values[2].(time.Duration).Nanoseconds()/1e4)/100.0))

			// sql
			vs := values[4].([]interface{})
			for _, value := range vs {
				indirectValue := reflect.Indirect(reflect.ValueOf(value))
				if indirectValue.IsValid() {
					value = indirectValue.Interface()
					if t, ok := value.(time.Time); ok {
						formattedValues = append(formattedValues, fmt.Sprintf("'%v'", t.Format("2006-01-02 15:04:05")))
					} else if b, ok := value.([]byte); ok {
						if str := string(b); isPrintable(str) {
							formattedValues = append(formattedValues, fmt.Sprintf("'%v'", str))
						} else {
							formattedValues = append(formattedValues, "'<binary>'")
						}
					} else if r, ok := value.(driver.Valuer); ok {
						if value, err := r.Value(); err == nil && value != nil {
							formattedValues = append(formattedValues, fmt.Sprintf("'%v'", value))
						} else {
							formattedValues = append(formattedValues, "NULL")
						}
					} else {
						switch value.(type) {
						case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
							formattedValues = append(formattedValues, fmt.Sprintf("%v", value))
							break
						default:
							formattedValues = append(formattedValues, fmt.Sprintf("'%v'", value))
						}
					}
				} else {
					formattedValues = append(formattedValues, "NULL")
				}
			}

			// differentiate between $n placeholders or else treat like ?
			if gromNumericPlaceHolderRegexp.MatchString(values[3].(string)) {
				sql = values[3].(string)
				for index, value := range formattedValues {
					placeholder := fmt.Sprintf(`\$%d([^\d]|$)`, index+1)
					sql = regexp.MustCompile(placeholder).ReplaceAllString(sql, value+"$1")
				}
			} else {
				formattedValuesLength := len(formattedValues)
				vss := gromSQLRegexp.Split(values[3].(string), -1)
				for index, value := range vss {
					sql += value
					if index < formattedValuesLength {
						sql += formattedValues[index]
					}
				}
			}
			messages = append(messages, sql)
			messages = append(messages, fmt.Sprintf(" \n\033[36;31m[%v]\033[0m ", strconv.FormatInt(values[5].(int64), 10)+" rows affected or returned "))

		} else {
			messages = append(messages, "\033[31;1m")
			messages = append(messages, values[2:]...)
			messages = append(messages, "\033[0m")
		}
	}
	return
}

//grom logFormatter of production
var gromLogFormatter = func(values ...interface{}) (messages []zap.Field, levels string) {
	if len(values) > 1 {
		var level = values[0]
		if level == "sql" && !logger.logConf.LogSelect {
			s, ok := values[3].(string)
			if ok && s != "" && strings.Index(s, "SELECT") >= 0 {
				return
			}
		}
		var (
			sql             string
			formattedValues []string
			currentTime     = ""
			source          = ""
		)
		currentTime = time.Now().Format("2006-01-02 15:04:05")
		source, _ = values[1].(string)

		messages = []zap.Field{}
		messages = append(messages, zap.String("source", source))
		messages = append(messages, zap.String("time", currentTime))

		if level == "sql" {
			levels = "sql"
			cost := strconv.FormatFloat(float64(values[2].(time.Duration).Nanoseconds()/1e4)/100.0, 'f', 0, 64) + "ms"
			messages = append(messages, zap.String("cost", cost))

			// sql
			vs := values[4].([]interface{})
			for _, value := range vs {
				indirectValue := reflect.Indirect(reflect.ValueOf(value))
				if indirectValue.IsValid() {
					value = indirectValue.Interface()
					if t, ok := value.(time.Time); ok {
						formattedValues = append(formattedValues, fmt.Sprintf("'%v'", t.Format("2006-01-02 15:04:05")))
					} else if b, ok := value.([]byte); ok {
						if str := string(b); isPrintable(str) {
							formattedValues = append(formattedValues, fmt.Sprintf("'%v'", str))
						} else {
							formattedValues = append(formattedValues, "'<binary>'")
						}
					} else if r, ok := value.(driver.Valuer); ok {
						if value, err := r.Value(); err == nil && value != nil {
							formattedValues = append(formattedValues, fmt.Sprintf("'%v'", value))
						} else {
							formattedValues = append(formattedValues, "NULL")
						}
					} else {
						switch value.(type) {
						case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
							formattedValues = append(formattedValues, fmt.Sprintf("%v", value))
							break
						default:
							formattedValues = append(formattedValues, fmt.Sprintf("'%v'", value))
						}
					}
				} else {
					formattedValues = append(formattedValues, "NULL")
				}
			}
			// differentiate between $n placeholders or else treat like ?
			if gromNumericPlaceHolderRegexp.MatchString(values[3].(string)) {
				sql = values[3].(string)
				for index, value := range formattedValues {
					placeholder := fmt.Sprintf(`\$%d([^\d]|$)`, index+1)
					sql = regexp.MustCompile(placeholder).ReplaceAllString(sql, value+"$1")
				}
			} else {
				formattedValuesLength := len(formattedValues)
				vss := gromSQLRegexp.Split(values[3].(string), -1)
				for index, value := range vss {
					sql += value
					if index < formattedValuesLength {
						sql += formattedValues[index]
					}
				}
			}
			messages = append(messages, zap.String("sql", sql))
			rowsAffected := values[5].(int64)
			messages = append(messages, zap.Int64("rows", rowsAffected))
		} else {
			key, ok := level.(string)
			if !ok {
				key = "log"
			}
			levels = key
			messages = append(messages, zap.Any(key, values[2:]))
		}
	}
	return
}

func isPrintable(s string) bool {
	for _, r := range s {
		if !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}

func logLevel(text string) zapcore.Level {
	text = strings.ToLower(text)
	switch text {
	case "debug":
		return zapcore.DebugLevel
	case "info", "": // make the zero value useful
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "dpanic":
		return zapcore.DPanicLevel
	case "panic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	}
	return zapcore.ErrorLevel
}

//Clear clear all logs
func Clear() {
	Sync()
	logger = nil
	gromDebugLogger = nil
	gromLogFormatter = nil
}
