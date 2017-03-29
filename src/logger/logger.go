package logger

import (
    "fmt"
    "log"
)

const (
    KLogLevelDebug = iota
    KLogLevelInfo
    KLogLevelWarn
    KLogLevelError
    KLogLevelFatal
)

var logLevelPrefix = []string{
    "[DBG] ",
    "[INF] ",
    "[WRN] ",
    "[ERR] ",
    "[FAL]",
}

var iLogger ILogger

// Create ILogger Struct
type ILogger struct {
    // nothing
}

func (d *ILogger) Output(level int, callDepth int, f string) error {
    return log.Output(callDepth, logLevelPrefix[level] + "" + f)
}

func _log(level int, callDepth int, f string) {
    iLogger.Output(level, callDepth, f)
}

func LogDebug(f string, v ...interface{}) {
    _log(KLogLevelDebug, 2, fmt.Sprintf(f, v...))
}

func LogInfo(f string, v ...interface{}) {
    _log(KLogLevelInfo, 2, fmt.Sprintf(f, v...))
}

func LogWarn(f string, v ...interface{}) {
    _log(KLogLevelWarn, 2, fmt.Sprintf(f, v...))
}

func LogError(f string, v ...interface{}) {
    _log(KLogLevelError, 2, fmt.Sprintf(f, v...))
}

func LogFatal(f string, v ...interface{}) {
    _log(KLogLevelFatal, 2, fmt.Sprintf(f, v...))
}