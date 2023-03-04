package common

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"syscall"

	log "github.com/sirupsen/logrus"
)

func CheckPidIsRunning(pid int) error {

	process, err := os.FindProcess(pid)
	if err != nil {
		return NewError("Failed to find process: %s", err)
	} else {
		err := process.Signal(syscall.Signal(0))
		if err != nil {
			return NewError("pid %d already stopped", pid)
		}
		log.Debugf("pid %d is running", pid)
	}

	return nil

}

func NewError(format string, a ...interface{}) error {

	// print the error information
	var errorInfo string
	if a == nil {
		errorInfo = fmt.Sprintln(format)
	} else {
		errorInfo = fmt.Sprintf(format, a...)
	}

	log.Error(errorInfo)

	return errors.New(errorInfo)

}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func Cmder(name string, outputFlag bool, envs []string, arg ...string) bool {

	var result bool

	cmd := exec.Command(name, arg...)

	// set cmd parameter
	cmd.Stdin = os.Stdin
	if outputFlag {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if len(envs) != 0 {
		cmd.Env = append(os.Environ(), envs...)
	}

	if err := cmd.Run(); err != nil {
		log.Errorf("%s execute failed, %v ", name, err)
		result = false
	} else {
		log.Debugf("%s execute successfully. ", name)
		result = true
	}

	return result
}

func GetPersistentLogger(rootFolder string, logFileName string, logLevel log.Level) *log.Logger {

	if !PathExists(rootFolder) {
		if err := os.MkdirAll(rootFolder, 0644); err != nil {
			log.Errorf("GetPersistentLogger: failed to create folder %s . %v", rootFolder, err)
			return nil
		}
	}

	if logFileName == "" {
		logFileName = "logger.log"
	}

	ContainerRuntimeLog := path.Join(rootFolder, logFileName)
	logger, err := os.OpenFile(ContainerRuntimeLog, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		log.Errorf("GetPersistentLogger: failed to open file: %v\n", err)
	}

	return &log.Logger{
		Out:       io.MultiWriter(logger, os.Stdout),
		Level:     logLevel,
		Formatter: &log.JSONFormatter{},
	}
}
