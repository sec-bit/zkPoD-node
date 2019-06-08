package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

const (
	LshowBackGnd = 1 << iota
	LshowAttachedInfo
	LshowAttachedInfoSuffix
	LshortTextNotation
	LnoTextNotation
	LnoTime
	LnoPrefix
)

const (
	LOG_LEVEL_FATAL   = 1
	LOG_LEVEL_ERROR   = 3
	LOG_LEVEL_WARN    = 5
	LOG_LEVEL_INFO    = 7
	LOG_LEVEL_DEBUG   = 9
	LOG_LEVEL_DDEBUG  = 11
	LOG_LEVEL_DDDEBUG = 13
	LOG_LEVEL_BACKGND = 90

	FatalLevel   = 1
	ErrorLevel   = 3
	WarnLevel    = 5
	InfoLevel    = 7
	DebugLevel   = 9
	DDebugLevel  = 11
	DDDebugLevel = 13
	BackGndLevel = 90

	FORMAT_JSON = 1
	FORMAT_TEXT = 0
	JsonFormat  = 1
	TextFormat  = 0
)

type ILogger interface {
	SetLevel(debuglevel int)

	SetPrefix(prefix string)

	SetFormatter(format int)
	//	NewLogger(appendedPrefix string) *ILogger

	AddSessionInfo(key string, val string)

	SetFlags(flag int)

	Flags() int

	Println(v ...interface{})

	Printf(format string, v ...interface{})

	Backgndln(v ...interface{})

	Backgndf(format string, v ...interface{})

	Debugln(v ...interface{})

	Debugf(format string, v ...interface{})

	DDebugln(v ...interface{})

	DDebugf(format string, v ...interface{})

	DDDebugln(v ...interface{})

	DDDebugf(format string, v ...interface{})

	Infoln(v ...interface{})

	Infof(format string, v ...interface{})

	Warnln(v ...interface{})

	Warnf(format string, v ...interface{})

	Errorln(v ...interface{})

	Errorf(format string, v ...interface{})

	Fatalln(v ...interface{})

	Fatalf(format string, v ...interface{})
}

type NotationMap map[int]string

type Field struct {
	key string
	val string
}

type SimpleLogger struct {
	LogLevel     int
	flag         int
	FuncTop      string
	Prefix1      string
	AttachedInfo []Field
	Formatter    int
	Logger       *log.Logger
}

var Notations = NotationMap{
	LOG_LEVEL_BACKGND: "[\x1b[0;37mB\x1b[0m]",
	LOG_LEVEL_DDDEBUG: "[\x1b[0;36mDDD\x1b[0m]",
	LOG_LEVEL_DDEBUG:  "[\x1b[0;36mDD\x1b[0m]",
	LOG_LEVEL_DEBUG:   "[\x1b[0;36mD\x1b[0m]",
	LOG_LEVEL_INFO:    "[\x1b[1;32mI\x1b[0m]",
	LOG_LEVEL_WARN:    "[\x1b[1;33mW\x1b[0m]",
	LOG_LEVEL_ERROR:   "\x1b[1;31m[E]\x1b[0m",
	LOG_LEVEL_FATAL:   "\x1b[1;37;41m[F]\x1b[0m",
}

var LongTextNotations = NotationMap{
	LOG_LEVEL_BACKGND: "[\x1b[0;37mBACKG\x1b[0m]",
	LOG_LEVEL_DDDEBUG: "[\x1b[0;36mDDDEBUG\x1b[0m]",
	LOG_LEVEL_DDEBUG:  "[\x1b[0;36mDDEBUG\x1b[0m]",
	LOG_LEVEL_DEBUG:   "[\x1b[0;36mDEBUG\x1b[0m]",
	LOG_LEVEL_INFO:    "[\x1b[1;32mINFO \x1b[0m]",
	LOG_LEVEL_WARN:    "[\x1b[1;33mWARN \x1b[0m]",
	LOG_LEVEL_ERROR:   "\x1b[1;31m[ERROR]\x1b[0m",
	LOG_LEVEL_FATAL:   "\x1b[1;37;41m[FATAL]\x1b[0m",
}

var LevelString = NotationMap{
	LOG_LEVEL_BACKGND: "BKGND",
	LOG_LEVEL_DDDEBUG: "DEBUG3",
	LOG_LEVEL_DDEBUG:  "DEBUG2",
	LOG_LEVEL_DEBUG:   "DEBUG",
	LOG_LEVEL_INFO:    "INFO",
	LOG_LEVEL_WARN:    "WARN",
	LOG_LEVEL_ERROR:   "ERROR",
	LOG_LEVEL_FATAL:   "FATAL",
}

var stdLogger = log.New(os.Stdout, "", 0)

func GetCaller(depth int) string {

	// we get the callers as uintptrs - but we just need 1
	fpcs := make([]uintptr, 1)

	// skip 3 levels to get to the caller of whoever called Caller()
	n := runtime.Callers(3+depth, fpcs)
	if n == 0 {
		return "n/a" // proper error her would be better
	}

	// get the info of the actual function that's in the pointer
	fun := runtime.FuncForPC(fpcs[0] - 1)
	if fun == nil {
		return "n/a"
	}

	// return its name
	name := fun.Name()
	if len(name) <= 0 {
		return "void"
	}
	i := strings.LastIndex(name, ".")
	if i < 0 {
		return name
	}
	name = name[i+1 : len(name)]

	return name
}

func MyCaller() string {
	return GetCaller(1)
}

func GetCallChainUtilTop(topFunc string) string {

	stop := false
	callChain := ""
	i := 2
	for !stop {
		i++
		fname := GetCaller(i)
		if fname == "" {
			stop = true
			break
		}
		if fname == "n/a" {
			stop = true
			break
		}
		callChain = fname + "> " + callChain
		if fname == topFunc {
			stop = true
			break
		}
	}

	return callChain
}

// func (l *SimpleLogger) NewLogger(appendedPrefix string) *SimpleLogger {
// 	return &SimpleLogger{
// 		LogLevel:    l.LogLevel,
// 		flag:        l.flag,
// 		FuncTop:  "",
// 		Prefix1:     l.Prefix1 + " " + appendedPrefix,
// 		AttachedInfo: make([]Field, 0),
// 		Logger:      stdLogger,
// 		Formatter:   l.Formatter,
// 	}
// }

func (l *SimpleLogger) AddAttachedInfo(key string, val string) {
	l.AttachedInfo = append(l.AttachedInfo, Field{
		key: key,
		val: val,
	})
}

func (l *SimpleLogger) NewModuleLogger(module string) *SimpleLogger {
	var mod string
	if module == "" {
		mod = MyCaller()
	}
	return &SimpleLogger{
		LogLevel:     l.LogLevel,
		flag:         l.flag,
		FuncTop:      mod,
		Prefix1:      l.Prefix1,
		AttachedInfo: make([]Field, 0),
		Logger:       l.Logger,
		Formatter:    l.Formatter,
	}
}

func (l *SimpleLogger) AddModuleInfo(key string, val string) {
	l.AddAttachedInfo(key, val)
}

func (l *SimpleLogger) NewSessionLogger() *SimpleLogger {

	caller := MyCaller()
	return &SimpleLogger{
		LogLevel:     l.LogLevel,
		flag:         l.flag,
		FuncTop:      caller,
		Prefix1:      l.Prefix1,
		AttachedInfo: l.AttachedInfo,
		Logger:       l.Logger,
		Formatter:    l.Formatter,
	}
}

func (l *SimpleLogger) AddSessionInfo(key string, val string) {
	l.AddAttachedInfo(key, val)
}

func NewLogger(prefix string) *SimpleLogger {
	return &SimpleLogger{
		LogLevel:     LOG_LEVEL_DEBUG,
		flag:         0,
		FuncTop:      "",
		Prefix1:      prefix,
		AttachedInfo: make([]Field, 0),
		Logger:       stdLogger,
		Formatter:    TextFormat,
	}
}

func NewSimpleLogger(prefix string) *SimpleLogger {
	return &SimpleLogger{
		LogLevel:     LOG_LEVEL_DEBUG,
		flag:         0,
		FuncTop:      "",
		Prefix1:      prefix,
		AttachedInfo: make([]Field, 0),
		Logger:       stdLogger,
		Formatter:    TextFormat,
	}
}

func (l *SimpleLogger) SetFormatter(format int) {
	l.Formatter = format
	//	fmt.Println(" SetFormatter() ----------NOT IMPLEMENTED YET------------")
	return
}

func (l *SimpleLogger) SetLevel(level int) {
	if level > 0 {
		l.LogLevel = level
	}

	return
}

func (l *SimpleLogger) Flags() int {
	return l.flag
}

func (l *SimpleLogger) SetFlags(flag int) {
	l.flag = flag
	return
}

func (l *SimpleLogger) TurnOnBackGnd() {
	l.flag = l.flag | LshowBackGnd
	return
}

func (l *SimpleLogger) TurnOffBackGnd() {
	l.flag = l.flag &^ LshowBackGnd
	return
}

func (l *SimpleLogger) SetPrefix(prefix string) {
	l.Prefix1 = prefix
}

func (l *SimpleLogger) header(level int) string {
	var nt string
	var t string
	var prefix string
	var hd string
	if l.flag&LnoTextNotation != 0 {
		nt = ""
	} else if l.flag&LshortTextNotation != 0 {
		nt = Notations[level]
	} else {
		nt = LongTextNotations[level]
	}

	hd = nt

	if l.flag&LnoTime != 0 {
		t = ""
	} else {
		now := time.Now()
		year, month, day := now.Date()
		hour, min, sec := now.Clock()
		t = fmt.Sprintf("%d-%02d-%02d %d:%d:%d", year, month, day, hour, min, sec)
		hd = hd + " " + t
	}

	if l.flag&LnoPrefix != 0 {
		prefix = ""
	} else {
		prefix = "<" + l.Prefix1 + ">"
		hd = hd + " " + prefix
	}

	if l.flag&(LshowAttachedInfo) != 0 && (l.flag&(LshowAttachedInfoSuffix) == 0) {
		var sess string
		for _, f := range l.AttachedInfo {
			sess = sess + " [" + f.key + "=" + f.val + "] "
		}
		hd = hd + sess
	}

	if l.FuncTop == "" {
		return hd
	}
	// FuncTop = TopFunc
	callChain := GetCallChainUtilTop(l.FuncTop)
	hd = hd + " " + callChain

	return hd
}

func (l *SimpleLogger) footer(level int) string {
	var suffix string

	if l.flag&(LshowAttachedInfo) != 0 && (l.flag&(LshowAttachedInfoSuffix) != 0) {
		var sess string
		for _, f := range l.AttachedInfo {
			sess = sess + " [" + f.key + "=" + f.val + "] "
		}
		suffix = sess
	}

	return suffix
}

func (l *SimpleLogger) println(level int, v ...interface{}) {
	if l.LogLevel < level {
		return
	}
	if l.Formatter == TextFormat {
		l.Logger.Println(l.header(level) + fmt.Sprint(v...) + l.footer(level))
	} else if l.Formatter == JsonFormat {
		now := time.Now()
		year, month, day := now.Date()
		hour, min, sec := now.Clock()
		time := fmt.Sprintf("%d-%02d-%02d %d:%d:%d", year, month, day, hour, min, sec)
		content := []Field{Field{"timestamp", time},
			Field{"name", l.Prefix1},
			Field{"level", LevelString[level]},
		}
		if l.FuncTop != "" {
			content = append(content, Field{"module", l.FuncTop})
		}
		for _, si := range l.AttachedInfo {
			content = append(content, si)
		}
		content = append(content, Field{"message", fmt.Sprint(v...)})
		buf := &bytes.Buffer{}
		buf.Write([]byte{'{'})
		length := len(content)
		for i, c := range content {
			fmt.Fprintf(buf, "\"%s\": \"%v\"", c.key, c.val)
			if i < length-1 {
				buf.WriteByte(',')
			}
		}
		buf.Write([]byte{'}'})

		l.Logger.Println(buf.String())
	} else {
		fmt.Errorf("NOT-IMPLEMENTED-YET\n")
	}
}

func (l *SimpleLogger) printf(level int, format string, v ...interface{}) {
	if l.LogLevel < level {
		return
	}

	if l.Formatter == TextFormat {
		l.Logger.Printf(l.header(level) + fmt.Sprintf(format, v...) + l.footer(level))
	} else if l.Formatter == JsonFormat {
		now := time.Now()
		year, month, day := now.Date()
		hour, min, sec := now.Clock()
		time := fmt.Sprintf("%d-%02d-%02d %d:%d:%d", year, month, day, hour, min, sec)
		content := []Field{Field{"timestamp", time},
			Field{"name", l.Prefix1},
			Field{"level", LevelString[level]},
		}
		if l.FuncTop != "" {
			content = append(content, Field{"module", l.FuncTop})
		}
		for _, si := range l.AttachedInfo {
			content = append(content, si)
		}
		content = append(content, Field{"message", fmt.Sprintf(format, v...)})
		buf := &bytes.Buffer{}
		buf.Write([]byte{'{'})
		length := len(content)
		for i, c := range content {
			fmt.Fprintf(buf, "\"%s\": \"%v\"", c.key, c.val)
			if i < length-1 {
				buf.WriteByte(',')
			}
		}
		buf.Write([]byte{'}'})

		l.Logger.Println(buf.String())
	} else {
		fmt.Errorf("NOT-IMPLEMENTED-YET\n")
	}
}

func (l *SimpleLogger) Println(v ...interface{}) {
	l.println(LOG_LEVEL_DEBUG, v...)
}

func (l *SimpleLogger) Printf(format string, v ...interface{}) {
	l.printf(LOG_LEVEL_DEBUG, format, v...)
}

func (l *SimpleLogger) Backgndln(v ...interface{}) {
	if l.flag&LshowBackGnd != 0 {
		l.Logger.Println(l.header(LOG_LEVEL_BACKGND) + fmt.Sprint(v...))
	} else {
		l.println(LOG_LEVEL_BACKGND, v...)
	}
}

func (l *SimpleLogger) Backgndf(format string, v ...interface{}) {
	if l.flag&LshowBackGnd != 0 {
		l.Logger.Printf(l.header(LOG_LEVEL_BACKGND) + fmt.Sprint(v...))
	} else {
		l.printf(LOG_LEVEL_BACKGND, format, v...)
	}
}

func (l *SimpleLogger) DDDebugln(v ...interface{}) {
	l.println(LOG_LEVEL_DDDEBUG, v...)
}

func (l *SimpleLogger) DDDebugf(format string, v ...interface{}) {
	l.printf(LOG_LEVEL_DDDEBUG, format, v...)
}

func (l *SimpleLogger) DDebugln(v ...interface{}) {
	l.println(LOG_LEVEL_DDEBUG, v...)
}

func (l *SimpleLogger) DDebugf(format string, v ...interface{}) {
	l.printf(LOG_LEVEL_DDEBUG, format, v...)
}

func (l *SimpleLogger) Debugln(v ...interface{}) {
	l.println(LOG_LEVEL_DEBUG, v...)
}

func (l *SimpleLogger) Debugf(format string, v ...interface{}) {
	l.printf(LOG_LEVEL_DEBUG, format, v...)
}

func (l *SimpleLogger) Infoln(v ...interface{}) {
	l.println(LOG_LEVEL_INFO, v...)
}

func (l *SimpleLogger) Infof(format string, v ...interface{}) {
	l.printf(LOG_LEVEL_INFO, format, v...)
}

func (l *SimpleLogger) Warnln(v ...interface{}) {
	l.println(LOG_LEVEL_WARN, v...)
}

func (l *SimpleLogger) Warnf(format string, v ...interface{}) {
	l.printf(LOG_LEVEL_WARN, format, v...)
}

func (l *SimpleLogger) Errorln(v ...interface{}) {
	l.println(LOG_LEVEL_ERROR, v...)
}

func (l *SimpleLogger) Errorf(format string, v ...interface{}) {
	l.printf(LOG_LEVEL_ERROR, format, v...)
}

func (l *SimpleLogger) Fatalln(v ...interface{}) {
	l.println(LOG_LEVEL_FATAL, v...)
}

func (l *SimpleLogger) Fatalf(format string, v ...interface{}) {
	l.printf(LOG_LEVEL_FATAL, format, v...)
}
