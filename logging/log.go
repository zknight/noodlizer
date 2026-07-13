package logging

import (
	"fmt"
	"log"
	"log/syslog"
	"noodlizer/config"
	"os"
	"runtime"
	"strings"
)

/*
	type Log interface {
		Info(m string)
		Warning(m string)
		Error(m string)
		Debug(m string)
		Close() error
	}
*/
type lfunc func(...string)
type Logger struct {
	Error   lfunc
	Warning lfunc
	Info    lfunc
	Debug   lfunc
	slog    *syslog.Writer
	flog    *log.Logger
}

func NewLog(c config.Config) (*Logger, error) {
	if c.UseSyslog {
		sl, err := syslog.Dial("udp", c.SyslogHost, syslog.LOG_LOCAL6|syslog.LOG_DEBUG, "noodlizer")
		if err != nil {
			return nil, err
		}
		slog := &Logger{slog: sl}
		slog.Error = slog.sError
		slog.Warning = slog.sWarning
		slog.Info = slog.sInfo
		slog.Debug = slog.sDebug
		return slog, nil
	}

	f, err := os.OpenFile(c.LocalLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	flog := &Logger{flog: log.New(f, "noodlizer", log.Ldate|log.Ltime|log.Lshortfile)}
	flog.Error = flog.fError
	flog.Warning = flog.fWarning
	flog.Info = flog.fInfo
	flog.Debug = flog.fDebug
	return flog, nil

}

func addSrcLog(m string) string {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "???"
		line = 0
	}
	return fmt.Sprintf("<%s:%d> %s", file, line, m)
}
func (l *Logger) sInfo(m ...string) {
	l.slog.Info(addSrcLog(strings.Join(m, "-")))
}

func (l *Logger) sWarning(m ...string) {
	l.slog.Warning(addSrcLog(strings.Join(m, "-")))
}

func (l *Logger) sError(m ...string) {
	l.slog.Err(addSrcLog(strings.Join(m, "-")))
}

func (l *Logger) sDebug(m ...string) {
	l.slog.Debug(addSrcLog(strings.Join(m, "-")))
}

func (l *Logger) Close() error {
	if l.slog != nil {
		return l.slog.Close()
	}
	return nil
}

/*
func newFlog(p string) (*flog, error) {
	f, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	l := &flog{log: log.New(f, "noodlizer", log.Ldate|log.Ltime|log.Lshortfile)}
	return l, nil
}
*/

func (l *Logger) fInfo(m ...string) {
	l.output("info", strings.Join(m, "-"))
}

func (l *Logger) fWarning(m ...string) {
	l.output("warn", strings.Join(m, "-"))
}

func (l *Logger) fError(m ...string) {
	l.output("err", strings.Join(m, "-"))
}

func (l *Logger) fDebug(m ...string) {
	l.output("dbg", strings.Join(m, "-"))
}

func (l *Logger) output(lvl, m string) {
	t := fmt.Sprintf("[%s] %s", lvl, m)
	l.flog.Output(3, t)
}
