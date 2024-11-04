package logging

import (
	"fmt"
	"log"
	"os"
	"sort"
	"time"
)

type Logger struct {
	info, debug, err, warn *log.Logger
	fields                 map[string]string
}

func New() *Logger {
	return &Logger{
		info:   log.New(os.Stdout, "", 0),
		debug:  log.New(os.Stdout, "", 0),
		warn:   log.New(os.Stdout, "", 0),
		err:    log.New(os.Stdout, "", 0),
		fields: make(map[string]string, 0),
	}
}

func (l *Logger) WithFields(fields ...interface{}) *Logger {
	nl := New()
	nl.fields = mergeMaps(l.fields, createMap(fields))
	return nl
}

func (l *Logger) Info(msg string, fields ...interface{}) {
	l.log(msg, "INFO", l.info, fields)
}

func (l *Logger) Debug(msg string, fields ...interface{}) {
	l.log(msg, "DEBUG", l.debug, fields)
}

func (l *Logger) Error(err error, msg string, fields ...interface{}) {
	fields = append(fields, "error")
	fields = append(fields, err.Error())

	l.log(msg, "ERROR", l.err, fields)
}

func (l *Logger) log(msg, level string, il *log.Logger, fields ...interface{}) {
	date := time.Now().Format("2006-01-02 15:04:05")
	merged := mergeMaps(l.fields, createMap(fields))
	il.Printf("[%s] [%s] \"%s\" %s", date, level, msg, createString(merged))
}

func createString(fields map[string]string) string {
	keys := make([]string, len(fields))

	for k := range fields {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	rv := ""
	for _, k := range keys {
		if k == "" {
			continue
		}
		if len(rv) == 0 {
			rv = fmt.Sprintf("%s=%s", k, fields[k])
		} else {
			rv += fmt.Sprintf(" %s=%s", k, fields[k])
		}
	}
	return rv
}

func createMap(fields ...interface{}) map[string]string {
	ffields := flattenArgs(fields)

	rv := make(map[string]string, 0)
	if len(fields) == 0 {
		return rv
	}

	for i := 0; i < len(ffields); i += 2 {
		if i+1 <= len(ffields)-1 {
			rv[fmt.Sprint(ffields[i])] = addQuotes(ffields[i+1])
		} else {
			rv[fmt.Sprint(ffields[i])] = ""
		}
	}
	return rv
}

func mergeMaps(a, b map[string]string) map[string]string {
	rv := make(map[string]string, 0)

	for k, v := range a {
		rv[k] = v

	}

	for k, v := range b {
		rv[k] = v
	}

	return rv
}

func addQuotes(s any) string {
	switch s.(type) {
	case string:
		return fmt.Sprintf("\"%s\"", s)
	default:
		return fmt.Sprint(s)
	}
}

func flattenArgs(args ...interface{}) []interface{} {
	flattened := make([]interface{}, 0)

	for _, arg := range args {
		switch v := arg.(type) {
		case []interface{}:
			// If the argument is a slice, recursively flatten it.
			flattened = append(flattened, flattenArgs(v...)...)
		default:
			// Otherwise, append the argument to the flattened slice.
			flattened = append(flattened, arg)
		}
	}

	return flattened
}
