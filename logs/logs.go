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
	"runtime/debug"
	"strconv"
	"strings"
	"time"
	"unicode"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

var (
	logger                       *XLogger
	gromDebugLogger              = log.New(os.Stdout, "\r\n", 0)
	gromSQLRegexp                = regexp.MustCompile(`\?`)
	gromNumericPlaceHolderRegexp = regexp.MustCompile(`\$\d+`)
)

//LogConf log config
type LogConf struct {
	OutPath    string `json:"outPath"`
	Level      string `json:"level"`
	MaxSize    int    `json:"maxSize"`
	MaxBackups int    `json:"maxBackups"`
	MaxAge     int    `json:"maxAge"`
	Debug      bool   `json:"debug"`
	LogSelect  bool   `json:"logSelect"`
	Stdout     bool   `json:"-"`
}

//XLogger log
type XLogger struct {
	zap.Logger
	logConf *LogConf
}

//GormLogger Gorm loger
type GormLogger struct {
}

//Print Gorm log
func (*GormLogger) Print(v ...interface{}) {
	if logger.logConf.Stdout {
		msg := gromLogFormatterDebug(v...)
		if msg != nil {
			gromDebugLogger.Println(msg...)
		}
	} else {
		msg, _ := gromLogFormatter(v...)
		if msg != nil {
			source, _ := v[1].(string)
			InfoWithCaller("gorm", source, msg...)
		}
	}
}

//LogInit init log
func LogInit(conf *LogConf) *XLogger {
	if logger == nil {
		if conf == nil {
			conf = &LogConf{
				OutPath:    "./logs.log",
				Level:      "error",
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
		logger = &XLogger{Logger: *zap.New(core), logConf: conf}
	}
	return logger
}

//Info info log
func Info(msg string, fields ...zap.Field) {
	InfoWithCaller(msg, "", fields...)
}

//InfoWithCaller info log
func InfoWithCaller(msg string, caller string, fields ...zap.Field) {
	if logger != nil {
		logger.Info(msg, LogCaller(caller, 0, fields...)...)
	}
}

//Debug debug log
func Debug(msg string, fields ...zap.Field) {
	DebugWithCaller(msg, "", fields...)
}

//DebugWithCaller debug log
func DebugWithCaller(msg string, caller string, fields ...zap.Field) {
	if logger != nil {
		logger.Debug(msg, LogCaller(caller, 0, fields...)...)
	}
}

//Error error log
func Error(msg string, fields ...zap.Field) {
	ErrorWithCaller(msg, "", fields...)
}

//Errors Error log
func Errors(msg string, err error) {
	if err != nil {
		if msg != "" {
			msg += ","
		}
		msg += err.Error()
	}
	if msg != "" {
		ErrorWithCaller(msg, "")
	}
}

//ErrorWithCaller error log
func ErrorWithCaller(msg string, caller string, fields ...zap.Field) {
	if logger != nil {
		fields = LogCaller(caller, 0, fields...)
		if logger.logConf.Debug {
			fields = append(fields, zap.ByteString("stack", debug.Stack()))
		}
		logger.Error(msg, fields...)
	}
}

//Logger rew logger object
func Logger() *XLogger {
	return logger
}

//LoggerGorm Gorm logger object
func LoggerGorm() *GormLogger {
	return &GormLogger{}
}

//Sync sync log
func Sync() {
	if logger != nil {
		logger.Sync()
	}
}

func logLevel(text string) zapcore.Level {
	switch text {
	case "debug", "DEBUG":
		return zapcore.DebugLevel
	case "info", "INFO", "": // make the zero value useful
		return zapcore.InfoLevel
	case "warn", "WARN":
		return zapcore.WarnLevel
	case "error", "ERROR":
		return zapcore.ErrorLevel
	case "dpanic", "DPANIC":
		return zapcore.DPanicLevel
	case "panic", "PANIC":
		return zapcore.PanicLevel
	case "fatal", "FATAL":
		return zapcore.FatalLevel
	}
	return zapcore.ErrorLevel
}

//LogCaller callers
func LogCaller(caller string, skip int, fields ...zap.Field) []zap.Field {
	if caller == "" {
		if skip <= 0 {
			skip = 3
		}
		if caller == "" {
			caller = Caller(skip)
		}
	}
	if caller != "" {
		var fs []zap.Field
		if fields != nil {
			fs = fields[:]
		} else {
			fs = make([]zap.Field, 0, 1)
		}
		fs = append(fs, zap.String("caller", caller))
		return fs
	}
	return fields[:]
}

//Caller caller
func Caller(skip int) string {
	if skip <= 0 {
		skip = 3
	} else {
		skip++
	}
	_, file, line, ok := runtime.Caller(skip)
	if ok {
		caller := file + ":" + strconv.Itoa(line)
		return caller
	}
	return ""
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
			timeCost := strconv.FormatFloat(float64(values[2].(time.Duration).Nanoseconds()/1e4)/100.0, 'f', 0, 64) + "ms"
			messages = append(messages, zap.String("timeCost", timeCost))

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

//ClearLogger clear logs
func ClearLogger() {
	logger = nil
	gromDebugLogger = nil
	gromLogFormatter = nil
}
