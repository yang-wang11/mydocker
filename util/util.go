package util

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	log "github.com/Sirupsen/logrus"
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

// wrapper error
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

// check path
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
	cmd.Stdin  = os.Stdin
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