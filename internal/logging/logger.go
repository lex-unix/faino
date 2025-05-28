package logging

import (
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"

	"github.com/fatih/color"
)

type Level int

const (
	LevelDebug Level = -4
	LevelInfo  Level = 0
	LevelWarn  Level = 4
	LevelError Level = 8
)

var (
	debugColor = color.New(color.FgMagenta).SprintFunc()
	infoColor  = color.New(color.FgGreen).SprintFunc()
	warnColor  = color.New(color.FgYellow).SprintFunc()
	errorColor = color.New(color.FgRed).SprintFunc()
	blueColor  = color.New(color.FgBlue).SprintFunc()
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return ""
	}
}

func (l Level) ColorString() string {
	levelStr := l.String()
	switch l {
	case LevelDebug:
		return debugColor(levelStr)
	case LevelInfo:
		return infoColor(levelStr)
	case LevelWarn:
		return warnColor(levelStr)
	case LevelError:
		return errorColor(levelStr)
	default:
		return levelStr // return uncolored if unknown
	}
}

var defaultLogger atomic.Pointer[Logger]

func init() {
	defaultLogger.Store(New(os.Stdout, LevelInfo))
}

type Logger struct {
	out   io.Writer
	mu    sync.Mutex
	level Level
}

func Default() *Logger { return defaultLogger.Load() }

func SetDefault(l *Logger) { defaultLogger.Store(l) }

func New(out io.Writer, level Level) *Logger {
	return &Logger{
		out:   out,
		mu:    sync.Mutex{},
		level: level,
	}
}

func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

func Debug(msg string) {
	Default().logMessage(LevelDebug, msg)
}

func Info(msg string) {
	Default().logMessage(LevelInfo, msg)
}

func Warn(msg string) {
	Default().logMessage(LevelWarn, msg)
}

func Error(msg string) {
	Default().logMessage(LevelError, msg)
}

func DebugHost(host, msg string) {
	Default().logMessageWithHost(LevelDebug, host, msg)
}

func InfoHost(host, msg string) {
	Default().logMessageWithHost(LevelInfo, host, msg)
}

func WarnHost(host, msg string) {
	Default().logMessageWithHost(LevelWarn, host, msg)
}

func ErrorHost(host, msg string) {
	Default().logMessageWithHost(LevelError, host, msg)
}

func Debugf(format string, args ...any) {
	Default().logf(LevelDebug, format, args...)
}

func Infof(format string, args ...any) {
	Default().logf(LevelInfo, format, args...)
}

func Warnf(format string, args ...any) {
	Default().logf(LevelWarn, format, args...)
}

func Errorf(format string, args ...any) {
	Default().logf(LevelError, format, args...)
}

func DebugHostf(host, format string, args ...any) {
	Default().logfWithHost(LevelDebug, host, format, args...)
}

func InfoHostf(host, format string, args ...any) {
	Default().logfWithHost(LevelInfo, host, format, args...)
}

func WarnHostf(host, format string, args ...any) {
	Default().logfWithHost(LevelWarn, host, format, args...)
}

func ErrorHostf(host, format string, args ...any) {
	Default().logfWithHost(LevelError, host, format, args...)
}

// logMessage logs a simple message without formatting
func (l *Logger) logMessage(level Level, msg string) {
	if level < l.level {
		return
	}
	coloredLevel := level.ColorString()
	logLine := fmt.Sprintf("%s %s\n", coloredLevel, msg)
	l.mu.Lock()
	defer l.mu.Unlock()
	_, _ = l.out.Write([]byte(logLine))
}

// logf logs a formatted message
func (l *Logger) logf(level Level, format string, args ...any) {
	if level < l.level {
		return
	}
	formattedMsg := fmt.Sprintf(format, args...)
	coloredLevel := level.ColorString()
	logLine := fmt.Sprintf("%s %s\n", coloredLevel, formattedMsg)
	l.mu.Lock()
	defer l.mu.Unlock()
	_, _ = l.out.Write([]byte(logLine))
}

// logMessageWithHost logs a simple message with host prefix
func (l *Logger) logMessageWithHost(level Level, host string, msg string) {
	hostPart := fmt.Sprintf("[%s]", blueColor(host))
	fullMsg := fmt.Sprintf("%s %s", hostPart, msg)
	l.logMessage(level, fullMsg)
}

// logfWithHost logs a formatted message with host prefix
func (l *Logger) logfWithHost(level Level, host string, format string, args ...any) {
	formattedMsg := fmt.Sprintf(format, args...)
	hostPart := fmt.Sprintf("[%s]", blueColor(host))
	fullMsg := fmt.Sprintf("%s %s", hostPart, formattedMsg)
	l.logMessage(level, fullMsg)
}
