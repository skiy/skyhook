// Package skyhook is hook for sirupsen/logrus that used for writing the logs to local files.
package skyhook

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"sync"
)

// We are logging to file, strip colors to make the output more readable.
var defaultFormatter = &logrus.TextFormatter{}

// PathMap is map for mapping a log level to a file's path.
// Multiple levels may share a file, but multiple files may not be used for one level.
type PathMap map[logrus.Level]string

// WriterMap is map for mapping a log level to an io.Writer.
// Multiple levels may share a writer, but multiple writers may not be used for one level.
type WriterMap map[logrus.Level]io.Writer

// SkyHook is a hook to handle writing to local log files.
type SkyHook struct {
	paths     PathMap
	writers   WriterMap
	levels    []logrus.Level
	lock      *sync.Mutex
	formatter logrus.Formatter

	defaultPath      string
	defaultWriter    io.Writer
	hasDefaultPath   bool
	hasDefaultWriter bool
}

// NewHook returns new Sky hook.
// Output can be a string, io.Writer, WriterMap or PathMap.
// If using io.Writer or WriterMap, user is responsible for closing the used io.Writer.
func NewHook(output interface{}, formatter logrus.Formatter) *SkyHook {
	hook := &SkyHook{
		lock: new(sync.Mutex),
	}

	hook.SetFormatter(formatter)

	switch output.(type) {
	case string:
		hook.SetDefaultPath(output.(string))

	case io.Writer:
		hook.SetDefaultWriter(output.(io.Writer))

	case PathMap:
		hook.paths = output.(PathMap)
		for level := range hook.paths {
			hook.levels = append(hook.levels, level)
		}

	case WriterMap:
		hook.writers = output.(WriterMap)
		for level := range hook.writers {
			hook.levels = append(hook.levels, level)
		}

	default:
		panic(fmt.Sprintf("unsupported level map type: %v", reflect.TypeOf(output)))
	}

	return hook
}

// SetFormatter sets the format that will be used by hook.
// If using text formatter, this method will disable color output to make the log file more readable.
func (hook *SkyHook) SetFormatter(formatter logrus.Formatter) {
	hook.lock.Lock()
	defer hook.lock.Unlock()

	if formatter == nil {
		formatter = defaultFormatter
	}

	hook.formatter = formatter
}

// SetDefaultPath sets default path for levels that don't have any defined output path.
func (hook *SkyHook) SetDefaultPath(defaultPath string) {
	hook.lock.Lock()
	defer hook.lock.Unlock()

	hook.defaultPath = defaultPath
	hook.hasDefaultPath = true
}

// SetDefaultWriter sets default writer for levels that don't have any defined writer.
func (hook *SkyHook) SetDefaultWriter(defaultWriter io.Writer) {
	hook.lock.Lock()
	defer hook.lock.Unlock()

	hook.defaultWriter = defaultWriter
	hook.hasDefaultWriter = true
}

// Fire writes the log file to defined path or using the defined writer.
// User who run this function needs write permissions to the file or directory if the file does not yet exist.
func (hook *SkyHook) Fire(entry *logrus.Entry) error {
	hook.lock.Lock()
	defer hook.lock.Unlock()

	//log.SetOutput(ioutil.Discard)

	if hook.writers != nil || hook.hasDefaultWriter {
		return hook.ioWrite(entry)
	} else if hook.paths != nil || hook.hasDefaultPath {
		return hook.fileWrite(entry)
	}

	return nil
}

// Write a log line to an io.Writer.
func (hook *SkyHook) ioWrite(entry *logrus.Entry) error {
	var (
		writer io.Writer
		msg    []byte
		err    error
		ok     bool
	)

	if writer, ok = hook.writers[entry.Level]; !ok {
		if hook.hasDefaultWriter {
			writer = hook.defaultWriter
		} else {
			return nil
		}
	}

	// use our formatter instead of entry.String()
	msg, err = hook.formatter.Format(entry)

	if err != nil {
		log.Println("failed to generate string for entry:", err)
		return err
	}
	_, err = writer.Write(msg)
	return err
}

// Write a log line directly to a file.
func (hook *SkyHook) fileWrite(entry *logrus.Entry) error {
	var (
		fd   *os.File
		path string
		msg  []byte
		err  error
		ok   bool
	)

	if path, ok = hook.paths[entry.Level]; !ok {
		if hook.hasDefaultPath {
			path = hook.defaultPath
		} else {
			return nil
		}
	}

	logPath := filepath.Dir(path)
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		if err := os.MkdirAll(logPath, os.ModePerm); err != nil {
			log.Println("mkdir log path:", path, err)
			return err
		}
	}

	fd, err = os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Println("failed to open log file:", path, err)
		return err
	}
	defer fd.Close()

	// use our formatter instead of entry.String()
	msg, err = hook.formatter.Format(entry)

	if err != nil {
		log.Println("failed to generate string for entry:", err)
		return err
	}

	_, err = fd.Write(msg)
	return err
}

// Levels returns configured log levels.
func (hook *SkyHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
