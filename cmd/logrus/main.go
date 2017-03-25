package main

import (
	"fmt"
	"log/syslog"
	"os"
	"path"
	"runtime"
	"strconv"

	log "github.com/Sirupsen/logrus"
	logrus_syslog "github.com/Sirupsen/logrus/hooks/syslog"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

func isTrue(s string) bool {
	b, err := strconv.ParseBool(s)
	if err == nil && b {
		return true
	}
	i, err := strconv.ParseInt(s, 10, 8)
	if err == nil && i > 0 {
		return true
	}
	return false
}

func init() {
	logOutput := os.Stderr
	logLevel := log.InfoLevel

	for _, v := range []string{"GIT_TRACE", "DEBUG", "VERBOSE"} {
		if isTrue(os.Getenv(v)) {
			logLevel = log.DebugLevel
		}
	}

	hook, err := logrus_syslog.NewSyslogHook("", "", syslog.LOG_INFO, "")
	if err != nil {
		log.Error("Unable to connect to local syslog daemon")
	} else {
		log.AddHook(hook)
	}

	if !log.IsTerminal(logOutput) {
		log.SetFormatter(&log.TextFormatter{
			DisableColors:    true,
			DisableTimestamp: true,
		})
	} else {
		log.SetFormatter(&prefixed.TextFormatter{})
	}

	log.SetOutput(logOutput)
	log.SetLevel(logLevel)
}

func trace() string {
	pc := make([]uintptr, 10) // at least 1 entry needed
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	file, line := f.FileLine(pc[0])
	return fmt.Sprintf("%s:%d/%s", path.Base(file), line, f.Name())
}

type traceLogPrefix struct{}

func (p *traceLogPrefix) String() string {
	pc := make([]uintptr, 10) // at least 1 entry needed
	runtime.Callers(10, pc)
	f := runtime.FuncForPC(pc[0])
	file, line := f.FileLine(pc[0])
	return fmt.Sprintf("%s:%d/%s", path.Base(file), line, f.Name())
}

func main() {
	idLogger := log.WithFields(log.Fields{
		"uid": os.Getuid(),
		"gid": os.Getgid(),
		//"prefix": new(traceLogPrefix),
		"prefix": "main",
	})

	idLogger.Infof(`PATH="%s"`, os.Getenv("PATH"))
	idLogger.Printf(`PATH="%s"`, os.Getenv("PATH"))

	log.WithField("prefix", trace()).Warn("I see dead people.")
	idLogger.Debug("They're everywhere; walking around like regular people.")
	idLogger.WithField("prefix", trace()).Error("They don't know they're dead.")
}
