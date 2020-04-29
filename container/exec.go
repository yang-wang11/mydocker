package container

/*
#include <errno.h>
#include <sched.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>

__attribute__((constructor)) void injectProcess(void) {
//void injectProcess(void) {
	char *mydocker_pid;
	mydocker_pid = getenv("mydocker_pid");
	if (mydocker_pid) {
		//fprintf(stdout, "got mydocker_pid=%s\n", mydocker_pid);
	} else {
		fprintf(stdout, "missing mydocker_pid env skip");
		return;
	}
	char *mydocker_cmd;
	mydocker_cmd = getenv("mydocker_cmd");
	if (mydocker_cmd) {
		//fprintf(stdout, "got mydocker_cmd=%s\n", mydocker_cmd);
	} else {
		fprintf(stdout, "missing mydocker_cmd env skip");
		return;
	}
	int i;
	char nspath[1024];
	char *namespaces[] = { "ipc", "uts", "net", "pid", "mnt" };

	for (i=0; i<5; i++) {
		sprintf(nspath, "/proc/%s/ns/%s", mydocker_pid, namespaces[i]);
		int fd = open(nspath, O_RDONLY);

		if (setns(fd, 0) == -1) {
			fprintf(stderr, "setns on %s namespace failed: %s\n", namespaces[i], strerror(errno));
		} else {
			//fprintf(stdout, "setns on %s namespace succeeded\n", namespaces[i]);
		}
		close(fd);
	}
	int res = system(mydocker_cmd);
	exit(0);
	return;
}
*/
import "C"

import (
	. "docker/mydocker/util"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	log "github.com/Sirupsen/logrus"
)

const (
	dockerPid string = "mydocker_pid"
	dockerCmd string = "mydocker_cmd"
)

func ExecContainer(containerName string, command string) {

	// get container name
	con := GetContinerInfoByName(containerName)

	log.Debugf("Container pid %s, cmd %s", con.Pid, con.Command)

	// set environment
	os.Setenv(dockerPid, con.Pid)
	os.Setenv(dockerCmd, command)
	SetEnvPATH()

	// call
	Cmder("/proc/self/exe", true, GetEnvByPid(con.Pid), "exec")

}

func GetContinerInfoByName(containerName string) (con Container) {

	// read container content from file
	containerPath := path.Join(ContainerRuntimeBaseFolder, containerName)
	containerInfo, err := ioutil.ReadFile(path.Join(containerPath, ContainerConfigName))
	if err != nil {
		log.Errorf("read container %s info failed, %v", containerName, err)
		return
	}

	// parse container info
	err = json.Unmarshal(containerInfo, &con)
	if err != nil {
		log.Errorf("unmarshal container %s info failed, %v", containerName, err)
	}

	return

}

func GetEnvByPid(pid string) []string {

	pidEnvFile := fmt.Sprintf("/proc/%s/environ", pid)
	envContent, err := ioutil.ReadFile(pidEnvFile)
	if err != nil {
		log.Errorf("read file %s failed, %v", pidEnvFile, err)
	}

	envs := strings.Split(string(envContent), "\u0000")

	//log.Debugln("get env list", envs)

	return envs

}
